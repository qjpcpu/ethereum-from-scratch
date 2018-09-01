
# example

```
package main

import (
	"github.com/qjpcpu/filelog"
	"log"
)

func main() {
	w, err := filelog.NewWriter("test.log", filelog.RotateDaily, true)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte("hillo"))
	w.Close()
}
```
