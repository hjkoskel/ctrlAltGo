module webcamdemo

go 1.24.0

require (
	github.com/hjkoskel/ctrlaltgo v0.0.1
	github.com/hjkoskel/ctrlaltgo/initializing v0.0.0-20250305203652-709a906adb62
	github.com/peterhagelund/go-v4l2 v0.7.0
	puhveri v0.0.0

)

require golang.org/x/sys v0.28.0

replace puhveri => ../../puhveri

replace github.com/hjkoskel/ctrlaltgo/initializing => ../../initializing
