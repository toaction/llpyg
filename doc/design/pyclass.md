## Python Class

### Declaration

```Python
class Animal:
    def __init__(self, name):
        self._name = name

    def speak(self):
        pass

class Dog(Animal):

    dog_name = "Dog"

    def __init__(self, name, age):
        super().__init__(name)
        self._age = age

    def speak(self):
        print(f"Dog {self._name} is speaking")
    
    @property
    def age(self):
        return self._age

    @age.setter
    def age(self, age):
        self._age = age
    
    @classmethod
    def get_dog_name(cls):
        return cls.dog_name
    
    @staticmethod
    def get_dog_name_static():
        return "Dog"
    
    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"
```
- 实例方法第一个参数为 self，代表实例本身，不能省略
- 类方法第一个参数为 cls，代表类本身，不能省略
- 静态方法可以没有参数，与类或实例无关

存在以下符号信息：
- 类名称
- 继承关系
- 类属性
- 初始化方法
- 实例属性
- 实例方法
- 类方法
- 静态方法
- 特殊方法

### Usage

```Python
dog = Dog("Buddy", 10)

dog.speak()                 # instance method
dog.age                     # instance property
Dog.get_dog_name()          # class method
Dog.get_dog_name_static()   # static method
str(dog)                    # special method
```
- 类方法和静态方法既可以被类调用，也可以被实例调用
- property 被转为了属性，而不是方法
- 只有实现了 setter 方法的 property 才能直接赋值

## LLGo Bindings

### Class and Instance
LLGo Bindings: 对于没有父类的类，直接使用 `py.Object` 来表示。对于有父类的类，使用结构体来声明，将父类嵌入其中以实现继承能力。
```go
/* Animal class */
type Animal py.Object

//llgo:link NewAnimal py.Animal
func NewAnimal(name string) *Animal {return nil}


/* Dog class */
type Dog struct {
    Animal  // inherit  
}

//llgo:link NewDog py.Dog
func NewDog(name *py.Object, age *py.Object) *Dog {return nil}
```
use:
```go
dog := NewDog(py.Str("Buddy"), py.Long(10))
```
LLGo call:
```c
PyObject *module = PyImport_ImportModule("dog");
PyObject *dog_class = PyObject_GetAttrString(module, "Dog");
PyObject *args = PyTuple_New(2);
PyTuple_SetItem(args, 0, PyUnicode_FromString("Buddy"));
PyTuple_SetItem(args, 1, PyLong_FromLong(5));
PyObject *dog_instance = PyObject_CallObject(dog_class, args);
```
难点：
- cpython 返回的是 `py.Object`, 如何将其转为结构体？
- 当类有多个父类时，如何做类型映射？
- 当父类和子类不在同一模块时，如何定义父类？

### Methods
LLGo Bindings: 将实例方法、类方法、静态方法、特殊方法（去掉前后的双下划线）都统一转为结构体方法。
```go
//llgo:link (*Dog)Speak py.method.speak
func (d *Dog) Speak() *py.Object {return nil}
```
use:
```go
dog.Speak()
```
LLGo call:
```c
PyObject *result = PyObject_CallMethod(dog_instance, "speak", NULL);
```

### Attributes and Properties
LLGo Bindings: 将属性统一转为结构体方法, Get方法获取属性值, Set方法设置属性值。
```go
//llgo:link (*Dog)GetAge py.getAttr.age
func (d *Dog) GetAge() *py.Object {return nil}
```
use:
```go
age := dog.GetAge()
```
LLGo call:
```c
PyObject *result = PyObject_GetAttrString(dog_instance, "age");
```

