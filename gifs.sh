#!/bin/bash

# Runs the necessary commands to generate gifs for the README.md file.
cd ./examples/Button || exit
echo "Button"
vhs example.tape

cd ../ButtonGroup || exit
echo "ButtonGroup"
vhs example.tape

cd ../SearchableFilepicker || exit
echo "SearchableFilepicker"
vhs example.tape

cd ../TraversableFilepicker || exit
echo "TraversableFilepicker"
vhs example.tape
