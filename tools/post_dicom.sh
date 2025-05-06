#!/bin/sh
curl -sX POST --data-binary @./tools/data/example_eye.dcm http://localhost:8080/
