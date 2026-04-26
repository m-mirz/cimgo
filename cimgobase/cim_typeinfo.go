package cimgobase

type CIMTypeInfo struct {
	Id         string
	Label      string
	Namespace  string
	Origin     string
	Origins    []string
	Attributes map[string]CIMAttributeInfo
}

type CIMAttributeInfo struct {
	Id        string
	Label     string
	Namespace string
	Origin    string
}
