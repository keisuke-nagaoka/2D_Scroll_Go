package main

const (
	tileSize = 16

	screenWidth  = 320
	screenHeight = 192
	heightTiles  = screenHeight / tileSize

	gravity      = 0.5
	maxFallSpeed = 10.0
	moveSpeed    = 2.0
	jumpVelocity = -8.5

	playerW = 12
	playerH = 16

	enemyW     = 12
	enemyH     = 16
	enemySpeed = 1.0

	coinSize = 8

	oneUpSize  = 12
	oneUpSpeed = 1.2

	startingLives = 2

	numStages    = 10
	stageBannerFrames = 90 // ~1.5s at 60fps

	coyoteFrames      = 6
	jumpBufferFrames  = 6
)
