package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/**
// Ejemplo 1: Crear metalink básico
mg := NewMetalinkGenerator("Mi Serie", "Descripción", "Anime")
mg.AddMegaLink("https://mega.nz/#!abc123", "Mi_Serie", 1)
mg.SaveToFile("mi_serie.metalink")

// Ejemplo 2: Desde archivo de texto
err := ProcessFileToMetalink("enlaces.txt", "output.metalink")

// Ejemplo 3: Añadir archivos manualmente
mg := NewMetalinkGenerator("Colección", "Mis descargas", "Videos")
mg.AddFile("video1.mp4", "Primer video", "https://mega.nz/#!xyz789", 524288000, "mp4")
mg.SaveToFile("coleccion.metalink")`)
*/

// Estructuras para el formato Metalink XML

type Metalink struct {
	XMLName   xml.Name `xml:"metalink"`
	Xmlns     string   `xml:"xmlns,attr"`
	Generator string   `xml:"generator"`
	Published string   `xml:"published"`
	Files     []File   `xml:"file"`
}

type File struct {
	XMLName     xml.Name `xml:"file"`
	Name        string   `xml:"name,attr"`
	Description string   `xml:"description"`
	Size        int64    `xml:"size"`
	URLs        []URL    `xml:"url"`
	Hashes      []Hash   `xml:"hash,omitempty"`
}

type URL struct {
	XMLName    xml.Name `xml:"url"`
	Location   string   `xml:"location,attr,omitempty"`
	Preference int      `xml:"preference,attr,omitempty"`
	Value      string   `xml:",chardata"`
}

type Hash struct {
	XMLName xml.Name `xml:"hash"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",chardata"`
}

// MetalinkGenerator estructura principal del generador
type MetalinkGenerator struct {
	Title       string
	Description string
	Category    string
	Files       []MetalinkFile
}

// MetalinkFile representa un archivo individual
type MetalinkFile struct {
	Name        string
	Description string
	URL         string
	Size        int64
	Extension   string
}

// NewMetalinkGenerator crea un nuevo generador
func NewMetalinkGenerator(title, description, category string) *MetalinkGenerator {
	return &MetalinkGenerator{
		Title:       title,
		Description: description,
		Category:    category,
		Files:       make([]MetalinkFile, 0),
	}
}

// AddFile añade un archivo al metalink
func (mg *MetalinkGenerator) AddFile(name, description, url string, size int64, extension string) {
	file := MetalinkFile{
		Name:        name,
		Description: description,
		URL:         url,
		Size:        size,
		Extension:   extension,
	}
	mg.Files = append(mg.Files, file)
}

// AddMegaLink añade un enlace MEGA con detección automática
func (mg *MetalinkGenerator) AddMegaLink(url, baseName string, episodeNum int) {
	name := fmt.Sprintf("%s_Episodio_%02d.mkv", baseName, episodeNum)
	description := fmt.Sprintf("%s - Episodio %d", mg.Title, episodeNum)

	// Tamaño estimado para video anime (350MB por defecto)
	estimatedSize := int64(367001600)

	mg.AddFile(name, description, url, estimatedSize, "mkv")
}

// GenerateMetalink genera el contenido XML del metalink
func (mg *MetalinkGenerator) GenerateMetalink() (string, error) {
	metalink := Metalink{
		Xmlns:     "urn:ietf:params:xml:ns:metalink",
		Generator: "Go Metalink Generator v1.0",
		Published: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Files:     make([]File, 0),
	}

	// Convertir archivos internos al formato XML
	for _, file := range mg.Files {
		xmlFile := File{
			Name:        file.Name,
			Description: file.Description,
			Size:        file.Size,
			URLs: []URL{
				{
					Location:   "cloud",
					Preference: 100,
					Value:      file.URL,
				},
			},
		}
		metalink.Files = append(metalink.Files, xmlFile)
	}

	// Generar XML
	output, err := xml.MarshalIndent(metalink, "", "  ")
	if err != nil {
		return "", err
	}

	// Añadir declaración XML
	xmlDeclaration := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	return xmlDeclaration + string(output), nil
}

// SaveToFile guarda el metalink en un archivo
func (mg *MetalinkGenerator) SaveToFile(filename string) error {
	content, err := mg.GenerateMetalink()
	if err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// ExtractMegaLinks extrae enlaces de MEGA de un texto
//func ExtractMegaLinks(text string) []string {
//	megaRegex := regexp.MustCompile(`https://mega\.nz/#![A-Za-z0-9_-]+![A-Za-z0-9_-]+`)
//	return megaRegex.FindAllString(text, -1)
//}

// ParseEpisodeDocument parsea el documento de episodios específico
func ParseEpisodeDocument(text string) map[int]string {
	lines := strings.Split(text, "\n")
	episodes := make(map[int]string)
	currentEpisode := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detectar número de episodio
		if after, ok := strings.CutPrefix(line, "EPISODIO:"); ok {
			episodeStr := strings.TrimSpace(after)
			// Extraer número del episodio
			re := regexp.MustCompile(`Episodio (\d+)`)
			matches := re.FindStringSubmatch(episodeStr)
			if len(matches) > 1 {
				if num, err := strconv.Atoi(matches[1]); err == nil {
					currentEpisode = num
				}
			}
		}

		// Detectar enlace MEGA
		if strings.Contains(line, "mega.nz") && strings.HasPrefix(line, "Enlace:") && currentEpisode > 0 {
			url := strings.TrimSpace(strings.TrimPrefix(line, "Enlace:"))
			episodes[currentEpisode] = url
		}
	}

	return episodes
}

// CreateMetalinkFromText crea un metalink desde texto estructurado
func CreateMetalinkFromText(text, title, baseName string) (*MetalinkGenerator, error) {
	episodes := ParseEpisodeDocument(text)

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no se encontraron episodios con enlaces MEGA")
	}

	mg := NewMetalinkGenerator(title, "Serie completa de anime", "Anime")

	// Añadir episodios en orden
	for i := 1; i <= 12; i++ {
		if url, exists := episodes[i]; exists {
			mg.AddMegaLink(url, baseName, i)
		}
	}

	return mg, nil
}

// Función para procesar desde archivo
func ProcessFileToMetalink(inputFile, outputFile string) error {
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Detectar título del archivo
	title := strings.TrimSuffix(inputFile, ".txt")
	baseName := strings.ReplaceAll(title, " ", "_")

	mg, err := CreateMetalinkFromText(string(content), title, baseName)
	if err != nil {
		return err
	}

	return mg.SaveToFile(outputFile)
}

// Función utilitaria para validar URLs MEGA
func ValidateMegaURL(url string) bool {
	megaRegex := regexp.MustCompile(`^https://mega\.nz/#![A-Za-z0-9_-]+![A-Za-z0-9_-]+$`)
	return megaRegex.MatchString(strings.TrimSpace(url))
}

// Funciones adicionales para casos avanzados

// CreateMetalinkFromMegaList crea metalink desde lista simple de URLs
//func CreateMetalinkFromMegaList(urls []string, title, baseName string) *MetalinkGenerator {
//	mg := NewMetalinkGenerator(title, "Generado desde lista de URLs", "Downloads")
//
//	for i, url := range urls {
//		if ValidateMegaURL(url) {
//			mg.AddMegaLink(url, baseName, i+1)
//		}
//	}
//
//	return mg
//}

// AddMultipleURLsForFile añade múltiples URLs para el mismo archivo (mirrors)
//func (mg *MetalinkGenerator) AddMultipleURLsForFile(name, description string, urls []string, size int64) {
//	if len(urls) == 0 {
//		return
//	}
//
//	file := MetalinkFile{
//		Name:        name,
//		Description: description,
//		URL:         urls[0], // URL principal
//		Size:        size,
//		Extension:   getExtensionFromName(name),
//	}
//
//	mg.Files = append(mg.Files, file)
//
//	// TODO: Implementar soporte para múltiples URLs (mirrors) en la estructura XML
//}

// getExtensionFromName extrae la extensión de un nombre de archivo
//func getExtensionFromName(filename string) string {
//	parts := strings.Split(filename, ".")
//	if len(parts) > 1 {
//		return parts[len(parts)-1]
//	}
//	return "unknown"
//}

// BatchProcessFiles procesa múltiples archivos de enlaces
func BatchProcessFiles(inputFiles []string, outputDir string) error {
	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}
	}

	for _, inputFile := range inputFiles {
		// Generar nombre de salida
		baseName := strings.TrimSuffix(inputFile, ".txt")
		outputFile := fmt.Sprintf("%s.metalink", baseName)

		if outputDir != "" {
			outputFile = fmt.Sprintf("%s/%s", outputDir, outputFile)
		}

		// Procesar archivo
		err := ProcessFileToMetalink(inputFile, outputFile)
		if err != nil {
			fmt.Printf("⚠️  Error procesando %s: %v\n", inputFile, err)
			continue
		}

		fmt.Printf("✅ Procesado: %s -> %s\n", inputFile, outputFile)
	}

	return nil
}
