package uhost

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type instanceType struct {
	CPU           int
	Memory        int
	HostType      string
	HostScaleType string
}

func parseInstanceType(s string) (*instanceType, error) {
	split := strings.Split(s, "-")
	if len(split) < 3 {
		return nil, fmt.Errorf("instance type is invalid, got %q", s)
	}

	if split[1] == "customized" {
		return parseInstanceTypeByCustomize(split...)
	}

	return parseInstanceTypeByNormal(split...)
}
func (i *instanceType) String() string {
	if i.Iscustomized() {
		return fmt.Sprintf("%s-%s-%v-%v", i.HostType, i.HostScaleType, i.CPU, i.Memory)
	} else {
		return fmt.Sprintf("%s-%s-%v", i.HostType, i.HostScaleType, i.CPU)
	}
}

func (i *instanceType) Iscustomized() bool {
	return i.HostScaleType == "customized"
}

var instanceTypeScaleMap = map[string]int{
	"highcpu":  1 * 1024,
	"basic":    2 * 1024,
	"standard": 4 * 1024,
	"highmem":  8 * 1024,
}

var availableHostTypes = []string{"n"}

func parseInstanceTypeByCustomize(splited ...string) (*instanceType, error) {
	if len(splited) != 4 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-customize-1-2")
	}

	hostType := splited[0]
	err := checkStringIn(hostType, availableHostTypes)
	if err != nil {
		return nil, err
	}

	hostScaleType := splited[1]

	cpu, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}

	memory, err := strconv.Atoi(splited[3])
	if err != nil {
		return nil, fmt.Errorf("memory count is invalid, please use a number")
	}

	if cpu/memory > 2 || memory/cpu > 12 {
		return nil, fmt.Errorf("the ratio of cpu to memory should be range of 2:1 ~ 1:12, got %d:%d", cpu, memory)
	}

	if memory/cpu == 1 || memory/cpu == 2 || memory/cpu == 4 || memory/cpu == 8 {
		return nil, fmt.Errorf("instance type is invalid, expected %q like %q,"+
			"the Type can be highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8", "n-Type-CPU", "n-standard-1")
	}

	if cpu < 1 || 32 < cpu {
		return nil, fmt.Errorf("expected cpu to be in the range (1 - 32), got %d", cpu)
	}

	if memory < 1 || 128 < memory {
		return nil, fmt.Errorf("expected memory to be in the range (1 - 128),got %d", memory)
	}

	if cpu != 1 && (cpu%2) != 0 {
		return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
	}

	if memory != 1 && (memory%2) != 0 {
		return nil, fmt.Errorf("expected the number of memory must be divisible by 2 without a remainder (except single memory), got %d", memory)
	}

	t := &instanceType{}
	t.HostType = hostType
	t.HostScaleType = hostScaleType
	t.CPU = cpu
	t.Memory = memory * 1024
	return t, nil
}

var availableOutstandingCpu = []int{4, 8, 16, 32, 64}

func parseInstanceTypeByNormal(split ...string) (*instanceType, error) {
	if len(split) != 3 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	}

	hostType := split[0]
	err := checkStringIn(hostType, []string{"n", "o"})
	if err != nil {
		return nil, err
	}

	hostScaleType := split[1]

	if scale, ok := instanceTypeScaleMap[hostScaleType]; !ok {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	} else {
		cpu, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, fmt.Errorf("cpu count is invalid, please use a number")
		}

		if cpu != 1 && (cpu%2) != 0 {
			return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
		}

		if hostType == "o" {
			if err := checkIntIn(cpu, availableOutstandingCpu); err != nil {
				return nil, fmt.Errorf("expected cpu of outstanding instancetype %q", err)
			}

			if hostScaleType == "highmem" && cpu == 64 {
				return nil, fmt.Errorf("this instance type %q is not supported, please refer to instance type document", "o-highmem-64")
			}
		} else {
			if hostScaleType == "highmem" && cpu > 16 {
				return nil, fmt.Errorf("expected cpu to be in the range (1 - 16) for normal highmem instance type, got %d", cpu)
			}

			if cpu < 1 || 32 < cpu {
				return nil, fmt.Errorf("expected cpu to be in the range (1 - 32) for normal instance type, got %d", cpu)
			}
		}

		memory := cpu * scale

		t := &instanceType{}
		t.HostType = hostType
		t.HostScaleType = hostScaleType
		t.CPU = cpu
		t.Memory = memory
		return t, nil
	}
}

type imageInfo struct {
	ImageId   string
	ProjectId string
	Region    string
}

func (i *imageInfo) Id() string {
	return fmt.Sprintf("%s:%s", i.ProjectId, i.Region)
}

type imageInfoSet struct {
	m    map[string]imageInfo
	once sync.Once
}

func newImageInfoSet(vL []imageInfo) *imageInfoSet {
	s := imageInfoSet{}
	for _, v := range vL {
		s.Set(v)
	}
	return &s
}

func (i *imageInfoSet) init() {
	i.m = make(map[string]imageInfo)
}

func (i *imageInfoSet) Set(img imageInfo) {
	i.once.Do(i.init)

	i.m[img.Id()] = img
}

func (i *imageInfoSet) Remove(id string) {
	i.once.Do(i.init)

	delete(i.m, id)
}

func (i *imageInfoSet) Get(projectId, region string) *imageInfo {
	k := fmt.Sprintf("%s:%s", projectId, region)
	if v, ok := i.m[k]; ok {
		return &v
	}
	return nil
}

func (i *imageInfoSet) GetAll() []imageInfo {
	var vL []imageInfo
	for _, img := range i.m {
		vL = append(vL, img)
	}
	return vL
}
