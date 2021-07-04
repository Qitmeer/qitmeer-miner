#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

extern int meer_pow(char* pheader_str,int pheader_len,char* ptarget_str,uint8_t* result_nonce,uint8_t* end);
extern int end(uint8_t* end);