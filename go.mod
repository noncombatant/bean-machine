module noncombatant.org/bean-machine

go 1.18

require (
	github.com/pkg/xattr v0.4.1
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	golang.org/x/text v0.3.2
	id3 v0.0.0-00010101000000-000000000000
)

require (
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	golang.org/x/sys v0.0.0-20190412213103-97732733099d // indirect
)

replace id3 => ./id3
