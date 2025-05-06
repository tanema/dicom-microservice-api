package handlers

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/frame"
	"github.com/suyashkumar/dicom/pkg/tag"

	"github.com/tanema/dicom-microservice-api/app/middleware"
	"github.com/tanema/dicom-microservice-api/app/storage"
	"github.com/tanema/dicom-microservice-api/config"
)

const (
	iterMaxBatch    = 100
	frameBufferSize = 100
)

type (
	DICOMAPI struct {
		store storage.Storage
	}
	dicomDoc struct {
		Hash string `json:"hash"`
	}
	docListResponse struct {
		Docs []dicomDoc `json:"docs"`
	}
)

// Create the new API with access method for the DICOM docs.
func NewDICOMAPI(cfg *config.Config, store storage.Storage) http.Handler {
	api := &DICOMAPI{store: store}
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Path("/").Methods(http.MethodPost).HandlerFunc(api.newDoc)
	r.Path("/").Methods(http.MethodGet).HandlerFunc(api.listDocs)
	r.Path("/{hash}/info").Methods(http.MethodGet).HandlerFunc(api.docInfo)
	r.Path("/{hash}").Methods(http.MethodGet).HandlerFunc(api.viewDoc)
	return r
}

// POST create new doc in our storage system. The api will only accept valid
// dicom docs and will return the hash for the file so that it can be accessed at
// a later time. This hash will be the same for any duplicate files.
func (api *DICOMAPI) newDoc(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not read request body: %v", err))
		return
	}

	_, err = dicom.ParseUntilEOF(bytes.NewReader(data), nil)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not could not read DICOM document from request body: %v", err))
		return
	}

	hash, err := api.store.Store(r.Context(), bytes.NewReader(data))
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, fmt.Sprintf("could not store document: %v", err))
		return
	}

	jsonOk(w, dicomDoc{Hash: hash})
}

// GET / List all available dicom docs in the sytem. This is useful mostly for debugging
// reasons because we do not store filenames and the hash values are not really
// informative.
func (api *DICOMAPI) listDocs(w http.ResponseWriter, r *http.Request) {
	docs := docListResponse{Docs: []dicomDoc{}}
	err := api.store.Iterate(r.Context(), iterMaxBatch, func(hashes []string) error {
		for _, h := range hashes {
			docs.Docs = append(docs.Docs, dicomDoc{Hash: h})
		}
		return nil
	})
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, fmt.Sprintf("problem encountered while iterating document store: %v", err))
		return
	}
	jsonOk(w, docs)
}

// GET /{hash}/info get tag information from the specified dicom file. If no tag
// query parameter is specified, it will return json-formatted information on all
// available tags. If a specific tag is requested, then it will return a json
// object with only that info.
func (api *DICOMAPI) docInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	doc, err := api.store.Fetch(r.Context(), hash)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not fetch doc with hash %v : %v", hash, err))
		return
	}

	parsedDoc, err := dicom.ParseUntilEOF(doc, nil)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not could not read DICOM document: %v", err))
		return
	}

	response := map[string]any{}
	if rTag := r.URL.Query().Get("tag"); rTag != "" {
		info, err := tag.FindByName(r.URL.Query().Get("tag"))
		if err != nil {
			jsonErr(w, http.StatusBadRequest, fmt.Sprintf("invalid tag %v : %v", r.URL.Query().Get("tag"), err))
			return
		}

		element, err := parsedDoc.FindElementByTag(info.Tag)
		if err != nil {
			jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not find tag: %v : %v", info.Name, err))
			return
		}
		formatTags(response, []*dicom.Element{element})
	} else {
		formatTags(response, parsedDoc.Elements)
	}

	if err := doc.Close(); err != nil {
		log.Printf("[WARN] could not close dicom doc %v\n", err)
	}

	jsonOk(w, response)
}

func (api *DICOMAPI) viewDoc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	doc, err := api.store.Fetch(r.Context(), hash)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not fetch doc with hash %v : %v", hash, err))
		return
	}

	allFrames := []*frame.Frame{}
	frameChan := make(chan *frame.Frame, frameBufferSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for frame := range frameChan {
			allFrames = append(allFrames, frame)
		}
		wg.Done()
	}()

	_, err = dicom.ParseUntilEOF(doc, frameChan)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("could not could not read DICOM document: %v", err))
		return
	}
	wg.Wait()

	frameID, err := strconv.Atoi(r.URL.Query().Get("frame"))
	if err != nil {
		frameID = 0
	}

	if len(allFrames) == 0 {
		jsonErr(w, http.StatusBadRequest, "no image data in this doc")
		return
	} else if frameID >= len(allFrames) || frameID < 0 {
		jsonErr(w, http.StatusBadRequest, "frameID out of bounds")
		return
	}

	fr := allFrames[frameID]
	img, err := allFrames[frameID].GetImage()
	if err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("problem encountered while getting image: %v", err))
		return
	}
	buf := bytes.NewBuffer(nil)
	var contentType string
	if !fr.IsEncapsulated() {
		contentType = "image/png"
		if err := png.Encode(buf, img); err != nil {
			jsonErr(w, http.StatusBadRequest, fmt.Sprintf("problem encountered while encoding image: %v", err))
			return
		}
	} else {
		contentType = "image/jpeg"
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 100}); err != nil {
			jsonErr(w, http.StatusBadRequest, fmt.Sprintf("problem encountered while encoding image: %v", err))
			return
		}
	}

	if err := doc.Close(); err != nil {
		log.Printf("[WARN] could not close dicom doc %v\n", err)
	}

	w.Header().Add("Content-Type", contentType)
	_, _ = fmt.Fprint(w, buf.String())
}

func formatTags(resp map[string]any, elements []*dicom.Element) {
	for _, elem := range elements {
		info, _ := tag.Find(elem.Tag)
		if info.Name == "" {
			continue
		}

		if elem.Value.ValueType() == dicom.SequenceItem {
			data := map[string]any{}
			formatTags(data, elem.Value.GetValue().([]*dicom.Element))
			resp[info.Name] = data
		} else if elem.Value.ValueType() == dicom.Sequences {
			seq := elem.Value.GetValue().([]*dicom.SequenceItemValue)
			data := make([]map[string]any, len(seq))
			for i, item := range seq {
				data[i] = map[string]any{}
				formatTags(data[i], item.GetValue().([]*dicom.Element))
			}
			resp[info.Name] = data
		} else {
			resp[info.Name] = elem.Value.String()
		}
	}
}
