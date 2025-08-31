package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const urlBase = "https://www3.animeflv.net"

// Anime representa un anime encontrado en la búsqueda
type Anime struct {
	Name string
	Link string
}

// Episode representa un episodio de anime
type Episode struct {
	Name string
	Link string
}

// Download representa un enlace de descarga
type Download struct {
	ProviderName string
	DownloadURL  string
}

// searchAnime busca animes basado en el texto de búsqueda
func searchAnime(searchText string) ([]Anime, error) {
	var animesList []Anime

	// Construir URL con parámetros de búsqueda
	searchURL := fmt.Sprintf("%s/browse?q=%s", urlBase, url.QueryEscape(searchText))

	// Realizar petición HTTP GET
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error haciendo petición: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error accediendo a la página, código de estado: %d", resp.StatusCode)
	}

	// Parsear HTML con goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %v", err)
	}

	// Encontrar elementos específicos
	doc.Find(".ListAnimes .Anime").Each(func(i int, s *goquery.Selection) {
		animeLink := s.Find("a")
		animeName := animeLink.Find(".Title").Text()
		href, exists := animeLink.Attr("href")

		if exists && animeName != "" {
			animesList = append(animesList, Anime{
				Name: strings.TrimSpace(animeName),
				Link: href,
			})
		}
	})

	return animesList, nil
}

// getDownloadLinksEpisode obtiene los enlaces de descarga de un episodio específico
func getDownloadLinksEpisode(episodeLink string) ([]Download, error) {
	var downloadList []Download

	// Configurar chromedp con opciones más robustas
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-javascript", false), // Habilitamos JS ya que puede ser necesario
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-images", true),
		chromedp.Flag("disable-default-apps", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Configurar contexto con logging reducido
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	// Timeout más corto
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var htmlContent string

	// Intentar obtener el contenido con manejo de errores mejorado
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlBase+episodeLink),
		chromedp.Sleep(2*time.Second), // Esperar carga
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		// Si ChromeDP falla, intentar con HTTP simple
		return getDownloadLinksWithHTTP(episodeLink)
	}

	// Parsear el contenido HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %v", err)
	}

	// Obtener tabla de descargas
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 4 {
			providerName := strings.TrimSpace(tds.Eq(0).Text())
			downloadLink := tds.Eq(3).Find("a")
			href, exists := downloadLink.Attr("href")

			if exists && providerName != "" {
				downloadList = append(downloadList, Download{
					ProviderName: providerName,
					DownloadURL:  href,
				})
			}
		}
	})

	return downloadList, nil
}

// getDownloadLinksWithHTTP intenta obtener los enlaces usando solo HTTP (fallback)
func getDownloadLinksWithHTTP(episodeLink string) ([]Download, error) {
	var downloadList []Download

	// Crear cliente HTTP con headers que simulen un navegador real
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", urlBase+episodeLink, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	// Agregar headers para simular un navegador real
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error HTTP: código de estado %d", resp.StatusCode)
	}

	// Parsear HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML con HTTP: %v", err)
	}

	// Buscar tabla de descargas
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 4 {
			providerName := strings.TrimSpace(tds.Eq(0).Text())
			downloadLink := tds.Eq(3).Find("a")
			href, exists := downloadLink.Attr("href")

			if exists && providerName != "" {
				downloadList = append(downloadList, Download{
					ProviderName: providerName,
					DownloadURL:  href,
				})
			}
		}
	})

	return downloadList, nil
}

// sanitizeFilename limpia el nombre del archivo para que sea válido en el sistema de archivos
func sanitizeFilename(filename string) string {
	// Remover caracteres no válidos para nombres de archivo
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	cleaned := reg.ReplaceAllString(filename, "_")

	// Remover espacios extra y caracteres especiales
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.ReplaceAll(cleaned, " ", "_")

	// Limitar longitud del nombre
	if len(cleaned) > 100 {
		cleaned = cleaned[:100]
	}

	return cleaned
}

// writeDownloadsToFile escribe los enlaces de descarga a un archivo de texto
func writeDownloadsToFile(animeName string, episodes []Episode, allDownloads map[string][]Download) error {
	// Crear nombre de archivo limpio
	filename := sanitizeFilename(animeName) + ".txt"

	// Crear archivo
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creando archivo: %v", err)
	}
	defer file.Close()

	// Escribir encabezado
	fmt.Fprintf(file, "ENLACES DE DESCARGA - %s\n", animeName)
	fmt.Fprintf(file, "Generado el: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "========================================\n\n")

	// Escribir enlaces por episodio
	for _, episode := range episodes {
		downloads, exists := allDownloads[episode.Link]
		if !exists || len(downloads) == 0 {
			continue
		}

		fmt.Fprintf(file, "EPISODIO: %s\n", episode.Name)
		fmt.Fprintf(file, "----------------------------------------\n")

		for _, download := range downloads {
			fmt.Fprintf(file, "Proveedor: %s\n", download.ProviderName)
			fmt.Fprintf(file, "Enlace: %s\n\n", download.DownloadURL)
		}

		fmt.Fprintf(file, "\n")
	}

	ProcessFileToMetalink(filename, filename+".metalink")

	return nil
}

// getLinksEpisodes obtiene la lista de episodios de un anime
func getLinksEpisodes(animeName, animeLink string) ([]Episode, error) {
	fmt.Printf("Procesando: %s, %s\n\n", animeName, animeLink)

	var episodesList []Episode

	// Configurar chromedp con opciones mejoradas
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-javascript", false), // Habilitamos JS
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-images", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Configurar contexto sin logging para evitar errores de cookies
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...interface{}) {}))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlBase+animeLink),
		chromedp.Sleep(3*time.Second), // Esperar más tiempo para carga completa
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		// Si ChromeDP falla, intentar con HTTP simple
		return getEpisodesWithHTTP(animeLink)
	}

	// Parsear HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML: %v", err)
	}

	// Encontrar episodios
	doc.Find("ul.ListCaps li").Each(func(i int, s *goquery.Selection) {
		episodeLink := s.Find("a")
		episodeName := episodeLink.Find("p").Text()
		href, exists := episodeLink.Attr("href")

		if exists && episodeName != "" {
			episodesList = append(episodesList, Episode{
				Name: strings.TrimSpace(episodeName),
				Link: href,
			})
		}
	})

	if len(episodesList) == 0 {
		fmt.Println("Episodios no encontrados con ChromeDP, intentando con HTTP...")
		return getEpisodesWithHTTP(animeLink)
	}

	fmt.Printf("Total de episodios disponibles: %d\n\n", len(episodesList))
	return episodesList, nil
}

// getEpisodesWithHTTP obtiene episodios usando solo HTTP (fallback)
func getEpisodesWithHTTP(animeLink string) ([]Episode, error) {
	var episodesList []Episode

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", urlBase+animeLink, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	// Headers para simular navegador
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error HTTP: código de estado %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML con HTTP: %v", err)
	}

	// Buscar episodios
	doc.Find("ul.ListCaps li").Each(func(i int, s *goquery.Selection) {
		episodeLink := s.Find("a")
		episodeName := episodeLink.Find("p").Text()
		href, exists := episodeLink.Attr("href")

		if exists && episodeName != "" {
			episodesList = append(episodesList, Episode{
				Name: strings.TrimSpace(episodeName),
				Link: href,
			})
		}
	})

	if len(episodesList) == 0 {
		fmt.Println("Episodios no encontrados.")
	} else {
		fmt.Printf("Total de episodios disponibles: %d\n\n", len(episodesList))
	}

	return episodesList, nil
}

// processAnimes procesa la lista de animes y permite al usuario seleccionar uno
func processAnimes(animesList []Anime) error {
	fmt.Println("Lista de animes disponibles:")

	for i, anime := range animesList {
		fmt.Printf("%d.- Anime: %s, enlace: %s\n", i+1, anime.Name, anime.Link)
	}

	fmt.Print("\nSelecciona un número para generar archivo con enlaces de descarga: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error leyendo entrada: %v", err)
	}

	input = strings.TrimSpace(input)
	option, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("solo se aceptan números")
	}

	if option < 1 || option > len(animesList) {
		return fmt.Errorf("opción inválida")
	}

	selectedAnime := animesList[option-1]
	fmt.Printf("Seleccionado: %s, %s\n", selectedAnime.Name, selectedAnime.Link)
	fmt.Println("\nProcesando episodios...")

	episodesList, err := getLinksEpisodes(selectedAnime.Name, selectedAnime.Link)
	if err != nil {
		return fmt.Errorf("error obteniendo episodios: %v", err)
	}

	if len(episodesList) == 0 {
		return fmt.Errorf("no se encontraron episodios para este anime")
	}

	// Mapa para almacenar todos los downloads
	allDownloads := make(map[string][]Download)

	fmt.Println("Obteniendo enlaces de descarga de todos los episodios...")

	// Obtener enlaces de cada episodio
	for i, episode := range episodesList {
		fmt.Printf("Procesando episodio %d/%d: %s", i+1, len(episodesList), episode.Name)

		downloadList, err := getDownloadLinksEpisode(episode.Link)
		if err != nil {
			fmt.Printf(" ❌ Error: %v\n", err)
			continue
		}

		allDownloads[episode.Link] = downloadList

		if len(downloadList) > 0 {
			fmt.Printf(" ✅ %d enlaces encontrados\n", len(downloadList))
		} else {
			fmt.Printf(" ⚠️  Sin enlaces\n")
		}

		// Pausa más corta entre requests
		time.Sleep(500 * time.Millisecond)
	}

	// Escribir todos los enlaces al archivo
	err = writeDownloadsToFile(selectedAnime.Name, episodesList, allDownloads)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo: %v", err)
	}

	filename := sanitizeFilename(selectedAnime.Name) + ".txt"
	absPath, _ := filepath.Abs(filename)

	fmt.Printf("\n✅ ¡Proceso completado!\n")
	fmt.Printf("📁 Archivo generado: %s\n", filename)
	fmt.Printf("📍 Ubicación completa: %s\n", absPath)

	// Mostrar estadísticas
	totalEpisodes := len(episodesList)
	totalLinks := 0
	processedEpisodes := 0

	for _, downloads := range allDownloads {
		if len(downloads) > 0 {
			processedEpisodes++
			totalLinks += len(downloads)
		}
	}

	fmt.Printf("\n📊 Estadísticas:\n")
	fmt.Printf("   • Total de episodios: %d\n", totalEpisodes)
	fmt.Printf("   • Episodios procesados: %d\n", processedEpisodes)
	fmt.Printf("   • Total de enlaces: %d\n", totalLinks)

	return nil
}

func main() {
	// Definir argumentos de línea de comandos
	search := flag.String("search", "", "Nombre del anime a buscar")
	searchShort := flag.String("s", "", "Nombre del anime a buscar (versión corta)")
	flag.Parse()

	// Usar el valor del argumento si existe
	searchTerm := *search
	if searchTerm == "" {
		searchTerm = *searchShort
	}

	if searchTerm == "" {
		fmt.Println("No se proporcionó término de búsqueda.")
		fmt.Println("Uso: ./programa --search \"nombre del anime\" o ./programa -s \"nombre del anime\"")
		return
	}

	animesList, err := searchAnime(searchTerm)
	if err != nil {
		log.Fatalf("Error buscando anime: %v", err)
	}

	if len(animesList) == 0 {
		fmt.Println("Anime no encontrado.")
		return
	}

	if err := processAnimes(animesList); err != nil {
		log.Fatalf("Error procesando animes: %v", err)
	}
}
