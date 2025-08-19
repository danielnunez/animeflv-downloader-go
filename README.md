# AnimeFLV Downloader

Un scraper de enlaces de descarga para AnimeFLV escrito en Go. Esta herramienta te permite buscar animes y generar archivos de texto con todos los enlaces de descarga organizados por episodio.

## 🚀 Características

- 🔍 Búsqueda de animes por nombre
- 📋 Lista interactiva para seleccionar anime
- 📁 Generación automática de archivos de texto con enlaces
- 🔄 Sistema de fallback robusto (ChromeDP + HTTP)
- ✅ Indicadores de progreso en tiempo real
- 📊 Estadísticas detalladas del proceso
- 🌐 Multiplataforma (Linux, macOS, Windows)

## 📋 Requisitos

- **Go 1.25 o superior**
- **Google Chrome** instalado en el sistema (para ChromeDP)
- **Conexión a internet**

### Instalación de Chrome por plataforma

**Linux (Ubuntu/Debian):**

```bash
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
sudo apt update
sudo apt install google-chrome-stable
```

**macOS:**

```bash
brew install --cask google-chrome
```

**Windows:**

Descargar desde: <https://www.google.com/chrome/>

## 🛠️ Instalación

### Opción 1: Descargar binarios precompilados

Ve a la sección [Releases](https://github.com/danielnunez/animeflv-downloader/releases) y descarga el binario para tu plataforma:

- **Linux**: `animeflv-downloader-linux-amd64`
- **macOS**: `animeflv-downloader-darwin-amd64`
- **Windows**: `animeflv-downloader-windows-amd64.exe`

### Opción 2: Compilar desde el código fuente

#### 1. Clonar el repositorio

```bash
git clone https://github.com/danielnunez/animeflv-downloader.git
cd animeflv-downloader
```

#### 2. Instalar dependencias

```bash
go mod tidy
```

#### 3. Compilar para tu plataforma actual

```bash
go build -o animeflv-downloader
```

#### 4. O usar el Makefile para compilar para múltiples plataformas

```bash
# Compilar para todas las plataformas
make all

# Compilar solo para Linux
make linux

# Compilar solo para macOS
make macos-amd64

# Compilar solo para Windows
make windows

# Limpiar archivos compilados
make clean
```

Los binarios se generarán en:

- `./bin/linux/animeflv-downloader`
- `./bin/macos/animeflv-downloader`
- `./bin/windows/animeflv-downloader.exe`

## 🎮 Uso

### Sintaxis básica

```bash
./animeflv-downloader --search "nombre del anime"
# o versión corta:
./animeflv-downloader -s "nombre del anime"
```

### Ejemplos

```bash
# Buscar Attack on Titan
./animeflv-downloader -s "Attack on Titan"

# Buscar Naruto
./animeflv-downloader --search "Naruto"

# Buscar One Piece
./animeflv-downloader -s "One Piece"
```

### Flujo de uso

1. **Ejecutar el comando** con el nombre del anime
2. **Seleccionar** el anime de la lista numerada
3. **Esperar** mientras se procesan todos los episodios
4. **Obtener** el archivo de texto generado con todos los enlaces

### Ejemplo de ejecución

```text
$ ./animeflv-downloader -s "Attack on Titan"

Lista de animes disponibles:

1.- Anime: Shingeki no Kyojin, enlace: /anime/shingeki-no-kyojin
2.- Anime: Shingeki no Kyojin Season 2, enlace: /anime/shingeki-no-kyojin-season-2
3.- Anime: Shingeki no Kyojin Season 3, enlace: /anime/shingeki-no-kyojin-season-3

Selecciona un número para generar archivo con enlaces de descarga: 1
Seleccionado: Shingeki no Kyojin, /anime/shingeki-no-kyojin

Procesando episodios...

Procesando: Shingeki no Kyojin, /anime/shingeki-no-kyojin

Total de episodios disponibles: 25

Obteniendo enlaces de descarga de todos los episodios...
Procesando episodio 1/25: Episodio 1 ✅ 4 enlaces encontrados
Procesando episodio 2/25: Episodio 2 ✅ 4 enlaces encontrados
...

✅ ¡Proceso completado!
📁 Archivo generado: Shingeki_no_Kyojin_enlaces_descarga.txt
📍 Ubicación completa: /home/usuario/Shingeki_no_Kyojin_enlaces_descarga.txt

📊 Estadísticas:
   • Total de episodios: 25
   • Episodios procesados: 25
   • Total de enlaces: 100
```

## 📁 Formato del archivo generado

El archivo de texto generado tiene el siguiente formato:

```text
ENLACES DE DESCARGA - Shingeki no Kyojin
Generado el: 2025-08-19 15:30:45
========================================

EPISODIO: Episodio 1
----------------------------------------
Proveedor: Mega
Enlace: https://mega.nz/file/abc123

Proveedor: MediaFire
Enlace: https://www.mediafire.com/file/def456

Proveedor: Google Drive
Enlace: https://drive.google.com/file/d/ghi789


EPISODIO: Episodio 2
----------------------------------------
Proveedor: Mega
Enlace: https://mega.nz/file/jkl012
...
```

## ⚙️ Configuración avanzada

### Variables de entorno

```bash
# Timeout personalizado para ChromeDP (en segundos)
export CHROMEDP_TIMEOUT=30

# Pausa entre requests (en milisegundos)
export REQUEST_DELAY=500
```

### Flags de compilación

```bash
# Compilación optimizada para producción
go build -ldflags="-s -w" -o animeflv-downloader

# Compilación con información de debug
go build -gcflags="all=-N -l" -o animeflv-downloader
```

## 🛠️ Desarrollo

### Estructura del proyecto

```text
animeflv-downloader/
├── main.go              # Código principal
├── go.mod               # Dependencias de Go
├── go.sum               # Checksums de dependencias
├── Makefile             # Scripts de compilación
├── README.md            # Este archivo
└── bin/                 # Binarios compilados
    ├── linux/
    ├── macos/
    └── windows/
```

### Dependencias principales

- **[goquery](https://github.com/PuerkitoBio/goquery)** - Parsing HTML (jQuery para Go)
- **[chromedp](https://github.com/chromedp/chromedp)** - Automatización de Chrome
- **[flag](https://pkg.go.dev/flag)** - Manejo de argumentos CLI

### Contribuir

1. **Fork** el proyecto
2. **Crea** una nueva rama (`git checkout -b feature/nueva-caracteristica`)
3. **Commit** tus cambios (`git commit -am 'Agregar nueva característica'`)
4. **Push** a la rama (`git push origin feature/nueva-caracteristica`)
5. **Abre** un Pull Request

## 🐛 Solución de problemas

### Chrome no encontrado

```bash
# Linux
which google-chrome
# Si no existe, instalar Chrome

# macOS
which "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# Windows
where chrome.exe
```

### Permisos de ejecución (Linux/macOS)

```bash
chmod +x animeflv-downloader
```

### Firewall/Antivirus

- Algunos antivirus pueden bloquear ChromeDP
- Agregar excepción para el ejecutable si es necesario

### Problemas de red

- Verificar conexión a internet
- Comprobar que AnimeFLV esté accesible
- Usar VPN si hay restricciones geográficas

## 📝 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo [LICENSE](LICENSE) para más detalles.

## ⚠️ Disclaimer

Esta herramienta es solo para uso educativo y personal. Respeta los términos de servicio de AnimeFLV y las leyes de derechos de autor de tu país. Los desarrolladores no se hacen responsables del uso indebido de esta herramienta.

## 🤝 Soporte

Si encuentras algún problema o tienes sugerencias:

1. **Issues**: [GitHub Issues](https://github.com/danielnunez/animeflv-downloader/issues)
2. **Discusiones**: [GitHub Discussions](https://github.com/danielnunez/animeflv-downloader/discussions)
3. **Email**: <dnunezse@gmail.com>

---

⭐ **¡Dale una estrella al proyecto si te fue útil!** ⭐
