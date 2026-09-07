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

	enemyJumpVelocity = -6.0
	enemyJumpInterval = 40  // frames between hops for jumper/landed-dropper enemies
	enemyThrowInterval = 100 // frames between throws for thrower enemies

	// dropperTriggerRange is how close (in px) the player must be, horizontally,
	// before a dormant enemyDropper starts falling, so the player actually
	// sees it drop instead of finding it already landed.
	dropperTriggerRange = 90

	projectileW     = 8
	projectileH     = 8
	projectileSpeed = 1.6

	coinSize = 8

	dropItemSize  = 12
	dropItemSpeed = 1.2

	particleCount = 16

	startingLives = 2

	numStages    = 10
	stageBannerFrames = 90 // ~1.5s at 60fps

	coyoteFrames      = 6
	jumpBufferFrames  = 6
)
