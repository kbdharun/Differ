package main

/*
 * 	License: GPL-3.0-or-later
 * 	Authors:
 * 		Mateus Melchiades <matbme@duck.com>
 * 	Copyright: 2023
 */

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/vanilla-os/differ/core"
	"github.com/vanilla-os/differ/core/handlers"
)

func setupRouter() (*gin.Engine, error) {
	// Initialize storage database
	err := core.InitStorage("test.db")
	if err != nil {
		return nil, errors.New("Failed to init storage: " + err.Error())
	}

	// Initialize cache
	err = core.InitCache()
	if err != nil {
		return nil, errors.New("Failed to init cache: " + err.Error())
	}

	// Fetches authentications from storage
	auths, err := core.FetchAuthorizations()
	if err != nil {
		return nil, errors.New("Failed to fetch authorizations from storage: " + err.Error())
	}

	// If auths is empty, we run the API in "read-only" mode.
	// In other words, we won't be able to add any images or releases via the API.
	readOnly := len(auths) == 0

	var authRequired gin.HandlerFunc
	if !readOnly {
		authRequired = gin.BasicAuth(auths)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Endpoint to check if API is running
	r.GET("/status", handlers.HandleStatus)

	// Manipulate images
	images := r.Group("/images")
	{
		// List all images
		images.GET("/", handlers.HandleGetImages)
		// List specific image
		images.GET("/:name", handlers.HandleFindImage)
		// Creates new image (Auth required)
		if !readOnly {
			images.POST("/new", authRequired, handlers.HandleAddImage)
		}

		// Release-related endpoints
		// Diffs two releases
		images.GET("/:name/diff", handlers.HandleGetReleaseDiff)
		// Gets latest release
		images.GET("/:name/latest", handlers.HandleGetLatestRelease)
		// Gets specific release with digest
		images.GET("/:name/:digest", handlers.HandleFindRelease)
		// Creates new release (Auth required)
		if !readOnly {
			images.POST("/:name/new", authRequired, handlers.HandleAddRelease)
		}
	}

	return r, nil
}

func main() {
	router, err := setupRouter()
	if err != nil {
		panic(err)
	}

	router.Run()
}
