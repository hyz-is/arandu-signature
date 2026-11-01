//go:build android

package native

/*
#include <jni.h>
#include <stdint.h>
#include <stdlib.h>

static int signature_open_url(uintptr_t vmHandle, uintptr_t ctxHandle, const char *url) {
	JavaVM *vm = (JavaVM *)vmHandle;
	jobject ctx = (jobject)ctxHandle;
	JNIEnv *env = NULL;
	int attached = 0;
	if ((*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6) != JNI_OK) {
		if ((*vm)->AttachCurrentThread(vm, &env, NULL) != JNI_OK) {
			return -1;
		}
		attached = 1;
	}
	int rc = -1;
	jclass uriClass = (*env)->FindClass(env, "android/net/Uri");
	jclass intentClass = (*env)->FindClass(env, "android/content/Intent");
	if (uriClass != NULL && intentClass != NULL) {
		jmethodID parse = (*env)->GetStaticMethodID(env, uriClass, "parse", "(Ljava/lang/String;)Landroid/net/Uri;");
		jmethodID construct = (*env)->GetMethodID(env, intentClass, "<init>", "(Ljava/lang/String;Landroid/net/Uri;)V");
		jmethodID addFlags = (*env)->GetMethodID(env, intentClass, "addFlags", "(I)Landroid/content/Intent;");
		jstring address = (*env)->NewStringUTF(env, url);
		jstring action = (*env)->NewStringUTF(env, "android.intent.action.VIEW");
		jobject uri = (*env)->CallStaticObjectMethod(env, uriClass, parse, address);
		jobject intent = (*env)->NewObject(env, intentClass, construct, action, uri);
		(*env)->CallObjectMethod(env, intent, addFlags, (jint)0x10000000);
		jclass contextClass = (*env)->GetObjectClass(env, ctx);
		jmethodID start = (*env)->GetMethodID(env, contextClass, "startActivity", "(Landroid/content/Intent;)V");
		(*env)->CallVoidMethod(env, ctx, start, intent);
		rc = 0;
	}
	if ((*env)->ExceptionCheck(env)) {
		(*env)->ExceptionClear(env);
		rc = -1;
	}
	if (attached) {
		(*vm)->DetachCurrentThread(vm);
	}
	return rc;
}
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/arandu-io/ayra/engine/app"
)

// OpenURL asks Android to open the address in the browser: an ACTION_VIEW
// intent started from the application's context, on a new task.
func OpenURL(url string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	vm, ctx := app.JavaVM(), app.AppContext()
	if vm == 0 || ctx == 0 {
		return errors.New("signature: the Android runtime is not ready")
	}
	address := C.CString(url)
	defer C.free(unsafe.Pointer(address))
	if C.signature_open_url(C.uintptr_t(vm), C.uintptr_t(ctx), address) != 0 {
		return errors.New("signature: Android refused to open " + url)
	}
	return nil
}
