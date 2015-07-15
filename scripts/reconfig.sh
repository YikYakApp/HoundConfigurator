#! /bin/bash

HOUNDDIR=/home/ubuntu/hound
CONF=config.json
PREVCONF=config.json-previous


cd $HOUNDDIR
cp $CONF $PREVCONF
./HoundConfigurator -org <foo> -user <bar> -token <baz> -excl excluded-repos.txt > $CONF

# bounce the server if there were repo changes
cmp -s $CONF $PREVCONF > /dev/null
if [ $? -eq 1 ]; then
	sudo restart hound
fi


