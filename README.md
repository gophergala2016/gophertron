GopherTron
==========

A 2D [Tron](https://en.wikipedia.org/wiki/Tron_(video_game)) clone in Go. Allows up to 4 players at once. 

# Install:
`go get github.com/gophergala2016/gophertron`

# Usage:
```
  -http string
    	http service address (default "localhost:8080")
  -prof
    	enables profiling with gom
```

* Controls: <kbd>up</kbd>, <kbd>down</kbd>, <kbd>left</kbd>, <kbd>right</kbd> (self explanatory)
* Recommended field dimensions: 100 x 100.


# Demo:

(*The game isn't this abrupt and boring, and allows players from multiple machines to compete, this is just a demo*)

![alt text](http://gcommer.com/i/gophertron.gif "thanks gcommer for the video")

# BUGS:

* Some tiny gaps between walls and players on collision aren't filled well.
* No score-keeping.
* All players have the same color.
* Lack of code documentation, since there was a single developer who didn't care to document the code.
* Lack of tests
