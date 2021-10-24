package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var Client *http.Client
var Version string
var wmarkSrc image.Image

// fetchAndResizeImage takes the url parameter then resizes it based on
// the query parameters
func fetchAndResizeImage(c *gin.Context) {
	url := c.Param("imageID")
	// see if the first character is a slash and then remove it, currently an issue with httprouter
	if strings.Split(url, "")[0] == "/" {
		url = strings.TrimPrefix(url, "/")
	}
	if url == "" {
		c.JSON(400, map[string]string{
			"error": "no url found",
		})
		return
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	resp, err := Client.Do(req)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	confReader := bytes.NewReader(b)
	imgReader := bytes.NewReader(b)
	conf, format, err := image.DecodeConfig(confReader)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	var (
		newHeight int
		newWidth  int
	)
	resizedHeightParam := c.Query("height")
	// assume the height and width stay the same if they aren't listed
	if resizedHeightParam == "" {
		newHeight = conf.Height
	} else {
		newHeight, err = strconv.Atoi(resizedHeightParam)
		if err != nil {
			c.JSON(400, map[string]string{
				"error": fmt.Sprintf("%s is not a valid height", resizedHeightParam),
			})
			return
		}
	}
	resizedWidthParam := c.Query("width")
	if resizedWidthParam == "" {
		newWidth = conf.Width
	} else {
		newWidth, err = strconv.Atoi(resizedWidthParam)
		if err != nil {
			c.JSON(400, map[string]string{
				"error": fmt.Sprintf("%s is not a valid width", resizedWidthParam),
			})
			return
		}
	}
	src, _, err := image.Decode(imgReader)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}
	dst := imaging.Resize(src, newWidth, newHeight, imaging.Lanczos)
	encoded := &bytes.Buffer{}

	ft, err := imaging.FormatFromExtension(format)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}
	// create the first file here
	err = imaging.Encode(encoded, dst, ft)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}
	// resize the watermark
	wmarkResize := imaging.Resize(wmarkSrc, newWidth/2, newHeight/2, imaging.Lanczos)
	// create the third image
	image3 := image.NewRGBA(dst.Bounds())
	// copy the resized first image here
	draw.Draw(image3, dst.Bounds(), dst, image.ZP, draw.Src)
	// copy the watermark over
	draw.Draw(image3, wmarkResize.Bounds(), wmarkResize, image.ZP, draw.Over)
	newEncoded := &bytes.Buffer{}
	err = imaging.Encode(newEncoded, image3, ft)
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}
	c.Writer.Header().Set("Content-Length", strconv.Itoa(newEncoded.Len()))
	_, err = c.Writer.Write(newEncoded.Bytes())
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}
	return
}

func healthcheck(c *gin.Context) {
	c.JSON(200, map[string]string{
		"status":  "ok",
		"version": Version,
	})
	return
}

func serveDoc(c *gin.Context) {
	b, _ := ioutil.ReadFile("./README.md")
	c.Header("Content-Type", "text/markdown")
	c.Writer.Write(b)
	return
}

func main() {

	wmark, err := os.Open("watermark.png")
	if err != nil {
		panic(err)
	}
	wmarkSrc, _ = imaging.Decode(wmark)
	Client = &http.Client{}
	Version = os.Getenv("VERSION")
	if Version == "" {
		Version = "local"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := gin.Default()
	app.GET("/resize/image/*imageID", fetchAndResizeImage)
	app.GET("/healthcheck", healthcheck)
	// redirect the base path to the readme file
	app.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/README.md")
	})
	app.GET("/README.md", serveDoc)
	panic(app.Run("0.0.0.0:" + port))
}
