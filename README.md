# PocketHealth Assignment
A microservice that is able to store, view, and query an uploaded DICOM file

### Usage

- `make serve` to run the application
- `make test` to run all tests
- `make help` to see any other tasks

Find the scripts to run endpoints
  - `./tools/post_dicom.sh` to create the doc on the server
  - `./tools/getdoc.sh` to view headers. You can edit this to view single tags by adding a tag query to the url
  - `./tools/list.sh` to view all docs stored in the service
  - `./tools/view_img.sh` to open a frame image of the dicom doc.

### Design
- I have setup a filestore that could easily be setup to send data to s3 as well, it simply needs the configuration
- All responses are json except the image responses which are jpeg or png, depending
  on how they are encoded in the dicom file.
- The service has also been setup so that it can have varying configuration depending on environment.
- Since this is supposed to run as a microservice, I am not going to save filenames
  and just deal with files by their hashes, expecting that the consuming services
  would handle managing filenames linked to the stored hashes.
- Hash as name is good for obscuring patient data, however a service like this would
  require a lot more security than I have provided here, such as only being accessible
  by private network, and only allowing images to be accessed through an authenticated
  proxy.

### Assignment
Design a RESTful API that is able to
  -[x] accept and store an uploaded DICOM file
  -[x] extract and return any DICOM header attribute based on a DICOM Tag as a query parameter
  -[x] finally convert the file into a PNG for browser-based viewing.

Note: You may assume that storage of DICOM files locally is acceptable for this assignment.

### Resources
- https://dicomiseasy.blogspot.com/2011/10/introduction-to-dicom-chapter-1.html
- https://www.dicomlibrary.com/dicom/dicom-tags/
- https://github.com/suyashkumar/dicom
