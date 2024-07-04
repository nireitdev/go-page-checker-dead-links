package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// Verifica que el link exista (=True)
func ping(url string) bool {
	cliente := http.Client{Timeout: 10 * time.Second}
	res, err := cliente.Get(url)
	if err != nil {
		return false
	}
	return res.StatusCode == http.StatusOK
}

func parser(link string) ([]string, error) {
	base_url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	urls := []string{}
	cliente := http.Client{Timeout: 10 * time.Second}
	resp, err := cliente.Get(link)
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if len(n.Attr) > 0 {
				t := n.Attr[0]
				if t.Key == "href" {
					www := t.Val
					if len(www) > 0 {
						u, _ := url.Parse(www)
						www = base_url.ResolveReference(u).String()
						urls = append(urls, www)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return urls, nil
}

// Imprime el exito o fracaso del chequeo del link
func info(result bool, url string, be_verbose bool) {
	if result == false {
		fmt.Printf("ERROR: %s \n", url)
	} else if be_verbose {
		fmt.Printf("OK   : %s \n", url)
	}
}

func main() {
	url_base := flag.String("h", "", "URL del sitio a escanear")
	hilos := flag.Int("t", 4, "Numero de Threads concurrentes")
	verbose := flag.Bool("v", false, "Verbose. Imprime todos los resultados")
	flag.Parse()
	if *url_base == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	starttime := time.Now()

	//Idea:
	// - Parseo el dominio y agrego a la lista los nuevos urls
	//	si url pertenece a la base_url => parseo
	//	caso contrario, es de otro dominio => solo ping para verificar su existencia
	buffer := make(chan string, 5000)
	visitados := map[string]bool{}
	buffer <- *url_base
	var wg sync.WaitGroup
	var working sync.WaitGroup
	var lock sync.RWMutex

	for i := 0; i < *hilos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for url := range buffer {
				working.Add(1)

				lock.Lock()
				_, ok := visitados[url]
				lock.Unlock()

				if ok {
					//el url ya se escaneo anteriomente
					working.Done()
					continue
				}

				//nuevo url a verificar:
				if (len(url) >= len(*url_base)) && (url[:len(*url_base)] == *url_base) {
					//el link pertenece al dominio original entonces parseo por mas links:
					new_urls, err := parser(url)
					result := err == nil

					lock.Lock()
					visitados[url] = result
					lock.Unlock()

					for _, u := range new_urls {
						buffer <- u
					}
					info(result, url, *verbose)
				} else {
					//el url es externo y entonces solo hago "ping" buscando el status 200
					result := ping(url)

					lock.Lock()
					visitados[url] = result
					lock.Unlock()

					info(result, url, *verbose)
				}
				working.Done()
			}
		}()

	}

	time.Sleep(5 * time.Second)
	working.Wait()
	close(buffer)
	wg.Wait()

	elapsed := time.Since(starttime)
	fmt.Printf("Links examinados: %d\n", len(visitados))
	fails := 0
	for _, f := range visitados {
		if !f {
			fails++
		}
	}
	fmt.Printf("Links rotos: %d\n", fails)
	fmt.Printf("Tiempo tomado: %s\n", elapsed)
}
