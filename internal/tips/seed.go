package tips

import "time"

func defaultSeed() int64 { return time.Now().UnixNano() }
