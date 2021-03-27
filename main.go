package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	args := os.Args[1]              //recupere l'url
	url := strings.Split(args, "/") // je split l'url
	fichier := url[len(url)-1]      //je recup la derniere valeur
	fmt.Println(fichier)            //test

	nbpart := 6 //nombre de téléchargement simultané
	head, _ := http.Head(args)
	header := head.Header
	clength := header.Get("Content-Length")
	conlength, _ := strconv.Atoi(clength)

	offset := conlength / nbpart
	// je teste pour avoir si le site accepte le téléchargement simultané
	if header.Get("Accept-Ranges") != "bytes" {
		panic("the file is not divisible")
	}
	client := http.Client{} //je crée un client pour gerer mes telechargements simultanés
	wg := sync.WaitGroup{}  //waitgroup pour attendre que tous mes téléchargements finissent

	for index := 0; index < nbpart; index++ {
		wg.Add(1)
		start := (index * offset) + 1
		if index == 0 {
			start = 0
		}
		end := (index + 1) * offset
		if index == nbpart-1 {
			end = conlength
		}
		name := fmt.Sprintf("part%d.part", index+1) //je stocke les différents fichiers en attendant que mes goroutines finissent

		go func() {
			part, _ := os.Create(name)
			req, _ := http.NewRequest("GET", args, nil)                     // crée un nouveau request pour chaque index
			req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end)) //specifie les bytes à téléchargés pour chaque part
			resp, _ := client.Do(req)
			body, _ := ioutil.ReadAll(resp.Body)

			part.Write(body)
			resp.Body.Close()
			part.Close()

			wg.Done()
		}()
	}
	wg.Wait()
	out, _ := os.Create(fichier)
	for i := 0; i < nbpart; i++ {
		name := fmt.Sprintf("part%d.part", i+1)
		body, _ := ioutil.ReadFile(name)

		if i == 0 {
			out.Write(body)
		} else {
			out.WriteAt(body, int64((offset*i)+1))
		}
		os.Remove(name)
	}

}
