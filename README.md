# intelquery
query multiple intelligence apis

## compile

### mac

GOOS=darwin go build -o intelquery intelquery

### aws

GOOS=linux go build -o intelquery intelquery

###

edit query name and query syntax in queries dir

## environment

export SHODAN_KEY=<key>
export URLSCAN_KEY=<key>
