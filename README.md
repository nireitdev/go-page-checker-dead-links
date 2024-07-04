# go-page-checker-dead-links
Simple tool to scan a site for dead o brokens links. 
Nothing fancy, just plain Go and standard packages.
Crawl a host o page and check for valid links. 

## To run:
```bash
go run main.go -h https://mysite.com 
```

Additional options:
-t : number of concurrents chekers
-v : be verbose

```bash
go run main.go -h https://mysite.com -t 10 -v
```

