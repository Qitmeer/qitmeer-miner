void search_circle(const unsigned int *arg_value,unsigned long length,unsigned int *ret);
void stop_solver(void *ppDetectInfo);
void init_solver(int deviceID,void **ppDetectInfo);
int run_solver(int deviceID,void *ppDetectInfo,char* header,int headerLen,unsigned int offset,unsigned int range,unsigned char* target,unsigned int *Nonce,unsigned int *CycleNonces,double *average);