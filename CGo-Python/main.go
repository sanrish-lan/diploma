package main

/*
#cgo pkg-config: python-3.14-embed
#define PY_SSIZE_T_CLEAN
#include <Python.h>
#include <stdlib.h>

static PyObject *pFuncProcess = NULL;

static int init_app() {
    Py_Initialize();

    PyRun_SimpleString("import sys\nif '.' not in sys.path: sys.path.insert(0, '.')\n");

    PyObject *pModule = PyImport_ImportModule("vdom_app");
    if (!pModule) {
        PyErr_Print();
        return 0;
    }

    pFuncProcess = PyObject_GetAttrString(pModule, "process_request");
    if (!pFuncProcess || !PyCallable_Check(pFuncProcess)) {
        PyErr_Print();
        return 0;
    }

    Py_DECREF(pModule);
    PyEval_SaveThread();
    return 1;
}

typedef struct {
    int status;
    char* body;
} response_t;

static response_t* call_vdom_request(const char* method, const char* path, 
                                     const char** header_keys, const char** header_vals, int header_count) {
    PyGILState_STATE gstate = PyGILState_Ensure();

    response_t* resp = (response_t*)malloc(sizeof(response_t));
    resp->status = 500;
    resp->body = NULL;

    PyObject* py_headers = PyDict_New();
    for (int i = 0; i < header_count; i++) {
        PyObject* val = PyUnicode_FromString(header_vals[i]);
        PyDict_SetItemString(py_headers, header_keys[i], val);
        Py_DECREF(val);
    }

    PyObject* req_dict = PyDict_New();
    PyDict_SetItemString(req_dict, "method", PyUnicode_FromString(method));
    PyDict_SetItemString(req_dict, "path", PyUnicode_FromString(path));
    PyDict_SetItemString(req_dict, "headers", py_headers);
    Py_DECREF(py_headers);

    PyObject* args = PyTuple_Pack(1, req_dict);
    PyObject* result = PyObject_CallObject(pFuncProcess, args);
    Py_DECREF(args);
    Py_DECREF(req_dict);

    if (result && PyDict_Check(result)) {
        PyObject* pStatus = PyDict_GetItemString(result, "status");
        if (pStatus && PyLong_Check(pStatus)) {
            resp->status = (int)PyLong_AsLong(pStatus);
        }

        PyObject* pBody = PyDict_GetItemString(result, "body");
        if (pBody && PyUnicode_Check(pBody)) {
            resp->body = strdup(PyUnicode_AsUTF8(pBody));
        }
        Py_DECREF(result);
    } else {
        PyErr_Print();
    }

    PyGILState_Release(gstate);
    return resp;
}
*/
import "C"
import (
	"fmt"
	"log"
	"net/http"
	"unsafe"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	cMethod := C.CString(r.Method)
	cPath := C.CString(r.URL.RequestURI())
	defer C.free(unsafe.Pointer(cMethod))
	defer C.free(unsafe.Pointer(cPath))

	hCount := len(r.Header)
	cKeys := make([]*C.char, 0, hCount)
	cVals := make([]*C.char, 0, hCount)

	for k, v := range r.Header {
		cK := C.CString(k)
		cV := C.CString(v[0])
		defer C.free(unsafe.Pointer(cK))
		defer C.free(unsafe.Pointer(cV))
		cKeys = append(cKeys, cK)
		cVals = append(cVals, cV)
	}

	var pKeys, pVals **C.char
	if hCount > 0 {
		pKeys = &cKeys[0]
		pVals = &cVals[0]
	}

	resp := C.call_vdom_request(cMethod, cPath, pKeys, pVals, C.int(hCount))
	if resp == nil || resp.body == nil {
		http.Error(w, "500 Internal Python Error", http.StatusInternalServerError)
		return
	}
	defer C.free(unsafe.Pointer(resp.body))
	defer C.free(unsafe.Pointer(resp))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(int(resp.status))
	w.Write([]byte(C.GoString(resp.body)))
}

func main() {
	if C.init_app() == 0 {
		log.Fatal("Ошибка инициализации Python или импорта vdom_app.py")
	}

	fmt.Println("[Go] Успешно загружен Python-модуль 'vdom_app'")
	fmt.Println("[Go] Сервер доступен на: http://localhost:8080")

	http.HandleFunc("/", httpHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
