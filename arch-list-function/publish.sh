#!/bin/bash

# create an OpenWhisk action, providing the parameters file
# to enable Cloudant access from our function
# *NOTE* you need to copy parameters.json.in to parameters.json
# and provide your credentials to your bound Bluemix-hosted
# Cloudant service

wsk action create archList archlist.js -P parameters.json

# if you have already created your action, after making code
# changes you can simply
# $ wsk action update archList archlist.js
