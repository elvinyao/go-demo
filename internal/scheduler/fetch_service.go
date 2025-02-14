package scheduler

import "log"

// FetchService 示例，用于周期性轮询外部数据源
type FetchService struct{}

func NewFetchService() *FetchService {
	return &FetchService{}
}

func (f *FetchService) FetchData() error {
	log.Println("[FetchService] Fetching data from external source... (demo)")
	// 示例逻辑: HTTP请求或读取文件
	return nil
}
