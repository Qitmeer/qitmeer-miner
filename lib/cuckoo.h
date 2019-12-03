void search_circle(const unsigned int *arg_value,unsigned long length,unsigned int *ret);
int cuda_search(int deviceID,unsigned char* header,unsigned int *isFind,unsigned int *Nonce,unsigned int *CycleNonces,double *average,void **ppDetectInfo,unsigned char* target);
void stop_solver(void *ppDetectInfo);