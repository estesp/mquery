#!/bin/bash

# BUILD
docker build -t estesp/mplookup .

# PUSH
docker push estesp/mplookup

# UPDATE
wsk action update mplatformQuery
