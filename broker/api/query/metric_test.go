package query

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/lindb/lindb/mock"
	"github.com/lindb/lindb/parallel"
	"github.com/lindb/lindb/series"

	"github.com/golang/mock/gomock"
)

func TestMetricAPI_Search(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	executorFactory := parallel.NewMockExecutorFactory(ctrl)
	api := NewMetricAPI(nil, nil, executorFactory, nil)

	// param error
	mock.DoRequest(t, &mock.HTTPHandler{
		Method:         http.MethodGet,
		URL:            "/broker/state",
		HandlerFunc:    api.Search,
		ExpectHTTPCode: 500,
	})

	// param error
	mock.DoRequest(t, &mock.HTTPHandler{
		Method:         http.MethodGet,
		URL:            "/broker/state?db=test",
		HandlerFunc:    api.Search,
		ExpectHTTPCode: 500,
	})

	exec := parallel.NewMockExecutor(ctrl)
	executorFactory.EXPECT().
		NewBrokerExecutor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(exec)
	exec.EXPECT().Execute().Return(nil)
	exec.EXPECT().Error().Return(fmt.Errorf("err")).MaxTimes(2)
	mock.DoRequest(t, &mock.HTTPHandler{
		Method:         http.MethodGet,
		URL:            "/broker/state?db=test&sql=select f from cpu",
		HandlerFunc:    api.Search,
		ExpectHTTPCode: 500,
	})

	ch := make(chan *series.TimeSeriesEvent)

	time.AfterFunc(10*time.Millisecond, func() {
		it := series.NewMockGroupedIterator(ctrl)
		it.EXPECT().HasNext().Return(false)
		ch <- &series.TimeSeriesEvent{
			Series: it,
		}
		close(ch)
	})

	executorFactory.EXPECT().
		NewBrokerExecutor(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(exec)
	exec.EXPECT().Execute().Return(ch)
	exec.EXPECT().Error().Return(nil)
	mock.DoRequest(t, &mock.HTTPHandler{
		Method:         http.MethodGet,
		URL:            "/broker/state?db=test&sql=select f from cpu",
		HandlerFunc:    api.Search,
		ExpectHTTPCode: 200,
	})
}
