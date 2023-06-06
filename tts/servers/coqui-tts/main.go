package main

// NOTE please use version 3.10 of python

// #cgo pkg-config: python3-embed
// #include <Python.h>
import "C"

import (
	"fmt"
)

type PyObject C.PyObject

// IncRef : https://docs.python.org/3/c-api/refcounting.html#c.Py_INCREF
func (pyObject *PyObject) IncRef() {
	C.Py_IncRef(toc(pyObject))
}

// DecRef : https://docs.python.org/3/c-api/refcounting.html#c.Py_DECREF
func (pyObject *PyObject) DecRef() {
	C.Py_DecRef(toc(pyObject))
}

// togo converts a *C.PyObject to a *PyObject
func togo(cobject *C.PyObject) *PyObject {
	return (*PyObject)(cobject)
}

func toc(object *PyObject) *C.PyObject {
	return (*C.PyObject)(object)
}

func main() {
	defer C.Py_Finalize()
	C.Py_Initialize()

	// dir, err := filepath.Abs(filepath.Dir("./pyspeech/__init__.py"))
	// if err != nil {
	// 	log.Fatalf("error finding python file %+v", err)
	// }

	// cinitModule := C.CString("import sys\nsys.path.append(\"" + dir + "\")")
	// C.PyRun_SimpleString(cinitModule)
	// C.free(unsafe.Pointer(cinitModule))

	// cmodule := C.CString("speech")
	// obj := togo(C.PyImport_ImportModule(cmodule))
	// C.free(unsafe.Pointer(cmodule))

	// defer obj.DecRef()

	fmt.Println(C.GoString(C.Py_GetVersion()))
}
