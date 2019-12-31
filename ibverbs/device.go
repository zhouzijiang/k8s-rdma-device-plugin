// +build linux

package ibverbs

// #cgo LDFLAGS: -libverbs
// #include <stdlib.h>
// #include <infiniband/verbs.h>
import "C"

import (
	"fmt"
	"unsafe"
)

func IbvGetDeviceList() ([]IbvDevice, error) {
	var ibvDevList []IbvDevice
	var c_num C.int
	var index C.uint8_t
	var c_ptrdevice *C.struct_ibv_device
	var c_deviceAttr C.struct_ibv_device_attr
	var c_portAttr C.struct_ibv_port_attr

	c_devList := C.ibv_get_device_list(&c_num)

	if c_devList == nil {
		return nil, fmt.Errorf("failed to get IB devices list")
	}

	ptrSize := unsafe.Sizeof(c_ptrdevice)
	ptr := uintptr(unsafe.Pointer(c_devList))
	for i := 0; i < int(c_num); i++ {
		c_ptrdevice = *(**C.struct_ibv_device)(unsafe.Pointer(ptr))
		if c_ptrdevice == nil {
			break
		}

		c_prtIbCtx := C.ibv_open_device(c_ptrdevice)
		if c_prtIbCtx == nil {
			ptr += ptrSize
			continue
		}

		ret := C.ibv_query_device(c_prtIbCtx, &c_deviceAttr)
		if ret < 0 {
			C.ibv_close_device(c_prtIbCtx)
			ptr += ptrSize
			continue
		}

		c_device := *c_ptrdevice

		for index = 1; index <= c_deviceAttr.phys_port_cnt; index++ {
			ret = C.ibv_query_port(c_prtIbCtx, index, &c_portAttr)
			if ret < 0 {
				continue
			}

			if c_portAttr.state == C.IBV_PORT_ACTIVE {
				device := IbvDevice{
					Name:       C.GoString(&c_device.name[0]),
					DevName:    C.GoString(&c_device.dev_name[0]),
					DevPath:    C.GoString(&c_device.dev_path[0]),
					IbvDevPath: C.GoString(&c_device.ibdev_path[0]),
				}
				ibvDevList = append(ibvDevList, device)
				break
			}
		}

		C.ibv_close_device(c_prtIbCtx)

		ptr += ptrSize
	}

	C.ibv_free_device_list(c_devList)

	return ibvDevList, nil
}
