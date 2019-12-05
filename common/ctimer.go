package common


// #include <unistd.h>
// //#include <errno.h>
// //int usleep(useconds_t usec);
import "C"

func Usleep(sec int)  {
	second := 1000*sec
	C.usleep((C.uint)(second))
}
