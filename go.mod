module load-generator

go 1.15

require (
	github.com/cheekybits/genny v1.0.1-0.20190611084615-df3d48aa411e // indirect
	github.com/chromedp/chromedp v0.5.3
	github.com/massimo-gollo/DASHpher v0.0.0-20220409200419-ff162d19a8b7
	github.com/massimo-gollo/godash v1.1.1
	github.com/pborman/getopt/v2 v2.1.0
	github.com/sirupsen/logrus v1.8.1
)

replace github.com/massimo-gollo/godash => ../godash

replace github.com/massimo-gollo/DASHpher => ../../DASHpher
