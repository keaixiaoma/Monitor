package collector

import (
	"context"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	v1 "github.com/RJuzhi/Monitor/api/v1"
	"github.com/RJuzhi/Monitor/pkg/log"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

type Collector struct {
	// CRD Info
	cache  cache.Cache
	client client.Client

	nodeName string

	// GPU Info
	cardList       v1.CardList
	cardNumber     uint
	FreeMemorySum  uint64
	TotalMemorySum uint64

	updateInterval int64
}

// countGPU count GPU nums
func (c *Collector) countGPU() {
	err := nvml.Init()
	if err != nil {
		log.ErrPrint(err)
	}
	defer func() {
		if err := nvml.Shutdown(); err != nil {
			log.ErrPrint(err)
		}
	}()
	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.ErrPrint(err)
	}
	c.cardNumber = count
}

func (c *Collector) updateGPU() {
	newCardList := make(v1.CardList, 0)
	err := nvml.Init()
	if err != nil {
		log.ErrPrint(err)
	}
	defer func() {
		if err := nvml.Shutdown(); err != nil {
			log.ErrPrint(err)
		}
	}()
	c.countGPU()
	for i := uint(0); i < c.cardNumber; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.ErrPrint(err)
		}
		health := "Healthy"
		status, err := device.Status()
		if err != nil {
			log.ErrPrint(err)
			health = "Unhealthy"
		}
		newCardList = append(newCardList, v1.Card{
			ID:          i,
			Health:      health,
			Model:       *device.Model,
			Power:       *device.Power,
			TotalMemory: *device.Memory,
			Clock:       *device.Clocks.Memory,
			FreeMemory:  *status.Memory.Global.Free,
			Core:        *device.Clocks.Cores,
			Bandwidth:   *device.PCI.Bandwidth,
		})
	}
	sort.Sort(newCardList)
	if len(c.cardList) == 0 || reflect.DeepEqual(c.cardList, newCardList) {
		c.cardList = newCardList
	}

	total, free := uint64(0), uint64(0)
	for _, card := range newCardList {
		total += card.TotalMemory
		free += card.FreeMemory
	}
	c.TotalMemorySum = total
	c.FreeMemorySum = free
	c.cardList = newCardList
}

func (c *Collector) Process() {
	interval := time.Duration(c.updateInterval) * time.Millisecond
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		c.updateGPU()

		currentMonitor := v1.Monitor{}

		key := types.NamespacedName{
			Name: c.nodeName,
		}

		err := c.client.Get(context.TODO(), key, &currentMonitor)
		if err != nil {
			log.ErrPrint(err)
			continue
		}
		// TODO update
		if c.NeedUpdate(currentMonitor.Status) {
			updateMonitor := currentMonitor.DeepCopy()
			updateMonitor.Status = v1.MonitorStatus{
				CardList:       c.cardList,
				CardNumber:     c.cardNumber,
				UpdateTime:     &metav1.Time{
					Time: time.Now(),
				},
				TotalMemorySum: c.TotalMemorySum,
				FreeMemorySum:  c.FreeMemorySum,
			}

			if err := c.client.Update(context.TODO(), updateMonitor); err != nil {
				log.ErrPrint(err)
			}
		}
	}
}

func NewCollector(interval int64, client client.Client, cache cache.Cache) *Collector {
	return &Collector{
		cardList: make(v1.CardList, 0),
		cardNumber: 0,
		updateInterval: interval,
		client: client,
		cache: cache,
	}
}

func StartCollector(c *Collector) {
	// Init CRD & Set Config
	c.nodeName = os.Getenv("NODE_NAME")
	// TODO create
	if err := c.createMonitor(); err != nil {
		panic(err)
	}
	c.Process()
}

func (c *Collector) createMonitor() error {
	monitor := v1.Monitor{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.nodeName,
		},
		Spec: v1.MonitorSpec{
			UpdateInterval: c.updateInterval,
		},
	}
	err := c.client.Create(context.TODO(), &monitor)
	if err != nil && !apierror.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (c *Collector) NeedUpdate(status v1.MonitorStatus) bool {
	if status.UpdateTime == nil {
		log.Print("CardList is Null, needs update.")
		return true
	}
	if status.TotalMemorySum != c.TotalMemorySum {
		log.Print("Total memory changed, needs update.")
		return true
	}
	if status.FreeMemorySum != c.FreeMemorySum {
		log.Print("Free memory changed, needs update.")
		return true
	}
	if status.CardNumber != c.cardNumber {
		log.Print("Card Number changed, needs update.")
		return true
	}
	if !reflect.DeepEqual(status.CardList, c.cardList) {
		log.Print("Card List changed, needs update.")
		return true
	}
	return false
}
