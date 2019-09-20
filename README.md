# intelquery
query multiple intelligence apis

## Compile

### mac

GOOS=darwin go build -o intelquery intelquery

### aws

GOOS=linux go build -o intelquery intelquery

## Queries

Edit query name and query syntax by modifying the query Json file in queries dir

## Environment Variables

export SHODAN_KEY=\<key>

export URLSCAN_KEY=\<key>
