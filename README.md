# HoundConfigurator
Simple scripts to configure [Etsy's hound](https://github.com/etsy/Hound)'s code search engine from a github org.

## Building
HoundConfigurator relies on [Godep](https://github.com/tools/godep) for vendoring.

To build, simply: 

    godep go install

## What's in here

### main.go
Generates the config file based on your org's github repos. Github org details are passed in through flags.

### excluded-repos.txt
List repo names here that you wish to exclude from hound. Comments (lines starting with #) and empty lines okay.

### scripts/reconfig.sh
Add this to your crontab to periodically kick-off HoundConfigurator. If changes in your github org are detected (eg, repos added or removed), it will restart the hound server and pick up the new config. 

### scripts/hound.conf
Trivial init file for hound server.

## Problems? Questions?
Get in touch: [Michael Morrissey](https://github.com/mgmyak)

