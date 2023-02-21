package prometheus

import (
	lContext "github.com/nyan233/littlerpc/core/common/context"
	"github.com/nyan233/littlerpc/core/middle/plugin"
	perror "github.com/nyan233/littlerpc/core/protocol/error"
	"github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"reflect"
)

type Exporter struct {
	plugin.AbstractServer
	traffic      *prometheus.CounterVec
	counter      *prometheus.CounterVec
	intervalTime *prometheus.HistogramVec
}

func NewServer(addr string) plugin.ServerPlugin {
	err := http.ListenAndServe(addr, promhttp.Handler())
	if err != nil {
		panic(err)
	}
	exp := new(Exporter)
	exp.traffic = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "traffic",
		Help: "服务的出入口流量统计",
	}, []string{"addr", "service", "type"})
	exp.counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "counter",
		Help: "服务调用次数统计",
	}, []string{"addr", "service", "type"})
	return exp
}

func (e *Exporter) Receive4S(pub *plugin.Context, msg *message.Message) perror.LErrorDesc {
	addr := lContext.CheckLocalAddr(pub.PluginContext)
	e.traffic.WithLabelValues(addr.String(), msg.GetServiceName(), "recv").Add(float64(msg.GetAndSetLength()))
	e.counter.WithLabelValues(addr.String(), msg.GetServiceName(), "call_all").Inc()
	return nil
}

func (e *Exporter) Call4S(pub *plugin.Context, args []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		addr := lContext.CheckLocalAddr(pub.PluginContext)
		service := lContext.CheckInitData(pub.PluginContext).ServiceName
		e.counter.WithLabelValues(addr.String(), service, "call_failed").Inc()
	}
	return nil
}

func (e *Exporter) AfterCall4S(pub *plugin.Context, args, results []reflect.Value, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		addr := lContext.CheckLocalAddr(pub.PluginContext)
		service := pub.PluginContext.Value("service").(string)
		e.counter.WithLabelValues(addr.String(), service, "call_failed").Inc()
	}
	return nil
}

func (e *Exporter) Send4S(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	if err != nil {
		addr := lContext.CheckLocalAddr(pub.PluginContext)
		service := lContext.CheckInitData(pub.PluginContext).ServiceName
		e.counter.WithLabelValues(addr.String(), service, "call_failed").Inc()
	}
	return nil
}

func (e *Exporter) AfterSend4S(pub *plugin.Context, msg *message.Message, err perror.LErrorDesc) perror.LErrorDesc {
	addr := lContext.CheckLocalAddr(pub.PluginContext)
	service := lContext.CheckInitData(pub.PluginContext).ServiceName
	if err != nil {
		e.counter.WithLabelValues(addr.String(), service, "call_failed").Inc()
		return nil
	}
	e.counter.WithLabelValues(addr.String(), service, "call_complete").Inc()
	e.traffic.WithLabelValues(addr.String(), service, "send").Add(float64(msg.GetAndSetLength()))
	return nil
}
