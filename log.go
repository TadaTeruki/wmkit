package wmkit

import (
	"log"
	"os"
)

func (sc *Screen) OpenLogAccess(filename string) {
	
	var err error
	sc.logFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("wmkit error: %v", err)
	}

	log.SetOutput(sc.logFile)
}

func (sc *Screen) CloseLogAccess() {
	sc.logFile.Close()
}
