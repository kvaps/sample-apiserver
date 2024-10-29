// pkg/serializer/custom_serializer.go

package serializer

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CustomSerializer оборачивает стандартный сериализатор и устанавливает правильный GVK.
type CustomSerializer struct {
	Delegate runtime.Serializer
}

// Decode декодирует данные и возвращает объект с установленным GVK.
func (cs *CustomSerializer) Decode(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	obj, gvk, err := cs.Delegate.Decode(data, defaults, into)
	if err != nil {
		return nil, nil, err
	}
	if obj != nil {
		obj.GetObjectKind().SetGroupVersionKind(*gvk)
	}
	return obj, gvk, nil
}

// Encode сериализует объект с учетом его GVK.
func (cs *CustomSerializer) Encode(obj runtime.Object, w io.Writer) error {
	return cs.Delegate.Encode(obj, w)
}

// SerializerFactory создает новый экземпляр CustomSerializer
func SerializerFactory(delegate runtime.Serializer) runtime.Serializer {
	return &CustomSerializer{
		Delegate: delegate,
	}
}

// Implement the Identifier method with the correct return type
func (s *CustomSerializer) Identifier() runtime.Identifier {
	return runtime.Identifier("your-custom-identifier") // Replace with a meaningful identifier
}
