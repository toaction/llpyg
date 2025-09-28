# 问题描述

目前，LLGo 仅支持调用 Python 函数，不支持处理类和对象。官方仓库 [goplus/lib/py](https://github.com/goplus/lib/py) 也缺少与 Python 类相关的 LLGo Bindings 代码示例。

为了扩展 LLGo 与 Python 生态系统的集成能力，需要添加对 Python 类的支持。一方面，llpyg 需要生成相应的 LLGo Bindings 代码，另一方面，LLGo 需要添加相应的处理逻辑。

## 符号信息

Python 类声明包含以下符号信息：
- 名称和继承关系
- 方法（包括类方法、实例方法、静态方法和特殊方法）
- 属性（包括类属性、实例属性）

Python 类示例：
```Python
class Animal:
    def __init__(self, name):
        self._name = name

    def speak(self):
        pass

class Dog(Animal):

    DOG_NAME = "Dog"

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
    def bark(cls, msg):
        print(f"Dog is barking {msg}")

    @staticmethod
    def sleep():
        print("Dog is sleeping")
    
    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"
```

使用示例：
```Python
dog = Dog("Buddy", 3)
dog.speak()
age = dog.age
dog.age = 4
Dog.DOG_NAME = "Buddy"
print(Dog.DOG_NAME)   # Buddy
Dog.bark("Woof")
Dog.sleep()
print(dog)
```

## LLGo Bindings 设计

### 设计原则

以下是 LLGo链接 Python 符号的核心代码：
```go
func (b Builder) pyLoadAttrs(obj Expr, name string) Expr {
	attrs := strings.Split(name, ".")
	fn := b.Pkg.pyFunc("PyObject_GetAttrString", b.Prog.tyGetAttrString())
	for _, attr := range attrs {
		obj = b.Call(fn, obj, b.CStr(attr))
	}
	return obj
}
```
LLGo 通过不断调用 `PyObject_GetAttrString` 来获取 Python 对象。基于此逻辑，LLGo Bindings 的设计旨在**方便用户使用的同时使 Go 符号能够正确链接到对应的 Python 符号**。


### 类
将 Python 类转换为 Go 结构体，通过嵌入结构体实现类似继承的功能。
> 目前结构体嵌入的方法仅支持单继承

例如 `Dog` 类继承自 `Animal` 类：
```Python
class Animal:
    pass

class Dog(Animal):
    pass
```
对应的 Go 结构体：
```Go
type Animal struct {
	py.Object
}

type Dog struct {
	Animal
}
```


### 实例

对于类实例，通过声明的构造函数 `New[ClassName]` 进行创建。

Python 代码示例：
```Python
class Dog(Animal):
    def __init__(self, name, age):
        super().__init__(name)
        self._age = age

dog = Dog("Buddy", 3)
```
符号链接规则：
```Python
attr = getattr(mod, "Dog")

dog = attr("Buddy", 3)
```
依据符号链接规则，构造函数链接的符号为 `py.[ClassName]`。

> 根据 LLGo 执行逻辑，通过链接 `py.[ClassName]` 来得到类实例实际上是先调用类的 `__new__` 方法，再调用 `__init__` 方法。Python 规定两个方法的参数必须一致。因此：
> - 当类显式声明了 `__init__` 方法时，构造函数的参数从 `__init__` 方法中获取。
> - 当类显式声明了 `__new__` 方法时，但未显式声明 `__init__` 方法时，构造函数的参数从 `__new__` 方法中获取。
> - 当两个方法都没有显式声明时，不提供该类的构造函数或提供空参的构造函数。


LLGo Bindings：

```Go
type Dog struct {
	Animal
}

//go:linkname NewDog py.Dog
func NewDog(name *py.Object, age *py.Object) *Dog
```

LLGo 使用方式：

```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
```

### 实例方法和魔法方法

对于实例方法和魔法方法，将其转为 Go 结构体方法。

Python 示例：

```Python
class Dog(Animal):
    def speak(self):
        print(f"Dog {self._name} is speaking")
    
    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"

dog = Dog("Buddy", 3)
dog.speak()
str(dog)
```
符号链接规则：
```Python
# 实例方法
attr = getattr(mod, "Dog")
attr = getattr(attr, "speak")
# call, 传入的第一个参数为实例对象
attr(dog)

# 魔法方法
attr = getattr(mod, "Dog")
attr = getattr(attr, "__str__")
# call, 传入的第一个参数为实例对象
attr(dog)
```
依据符号链接规则，实例方法和魔法方法链接的符号为 `py.[ClassName].[methodName]`。

LLGo Binding:
> 对于魔法方法，去除前后下划线，使其更符合命名规范与用户使用习惯。
```Go
//llgo:link (*Dog).Speak py.Dog.speak
func (d *Dog) Speak() *py.Object {
    return nil
}

//llgo:link (*Dog).Str py.Dog.__str__
func (d *Dog) Str() *py.Object {
    return nil
}
```
LLGo 使用方式：
```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
dog.Speak()
dog.Str()
```


### 实例属性
Python 示例：
```Python
class Dog(Animal):
    @property
    def age(self):
        return self._age

    @age.setter
    def age(self, age):
        self._age = age
```
```Python
dog = Dog("Buddy", 3)
dog.age
dog.age = 4
```
符号链接规则：
```Python
attr = getattr(mod, "Dog")
attr = getattr(attr, "age")
attr = getattr(attr, "__get__")
# call
age = attr(dog)

attr = getattr(mod, "Dog")
attr = getattr(attr, "age")
attr = getattr(attr, "__set__")
# call
attr(dog, 4)
```
依据符号链接规则，将实例属性的操作拆分为 `get` 和 `set` **结构体方法**，分别链接到符号 `py.[ClassName].[attributeName].__get__` 和 `py.[ClassName].[attributeName].__set__`。

LLGo Binding:
> 对于属性的获取操作，为了符合用户使用习惯，不添加 Get 前缀。

```Go
//llgo:link (*Dog).Age py.Dog.age.__get__
func (d *Dog) Age() *py.Object {
    return nil
}

//llgo:link (*Dog).SetAge py.Dog.age.__set__
func (d *Dog) SetAge(age *py.Object) {
}
```

LLGo 使用方式：

```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
dog.Age()
dog.SetAge(py.Long(4))
```

### 类属性
类属性属于类对象。在 Python 中可以直接通过类对象来获取和设置属性值。
```Python
class Dog(Animal):
    DOG_NAME = "Dog"

Dog.DOG_NAME
```
符号链接方式：和 Property 不同，类属性为具体的 Python 对象，对象字典中不存在类似于 `__get__` 和 `__set__` 的方法来获取和设置自身的值。

```Python
attr = getattr(mod, "Dog")
attr = getattr(attr, "DOG_NAME")
```

因此，无法通过方法调用的方式来获取和设置类属性。若想得到类属性，目前有两种处理方式：
1. 直接获取类属性，链接的符号为 `py.[ClassName].[AttributeName]`，此时获得的类属性是只读的。LLGo Binding 与类对应的结构体无关：
> 为了接近 Python 的语法以及防止命名冲突，添加 `[ClassName]` 作为前缀。
```go
//go:linkname DogDOGNAME py.Dog.DOG_NAME
var DogDOGNAME *py.Object
```
用户使用方式：
```go
dogName := DogDOGNAME
```
2. 先获取类对象，然后通过 `getAttribute` 和 `setAttribute` 方法获取和设置类属性，LLGo Binding 同样与类对应的结构体无关，且会声明一个同名的全局变量：
```go
//go:linkname Dog py.Dog
var Dog *py.Object
```
用户使用方式：
```go
dogClass := Dog
dogName := dogClass.getAttribute(py.Str("DOG_NAME"))
dogClass.setAttribute(py.Str("DOG_NAME"), py.Str("Buddy"))
```

### 类方法和静态方法
Python 示例：
```Python
class Dog:
    @classmethod
    def bark(cls, msg):
        print(f"Dog is barking {msg}")

    @staticmethod
    def sleep():
        print("Dog is sleeping")
```
符号链接规则：与实例方法不同，获取到类方法对象和静态方法对象后，可以直接调用，不应传入实例对象作为第一个参数。
```Python
# class method
attr = getattr(mod, "Dog")
attr = getattr(attr, "bark")
# call
attr("Hello")

# static method
attr = getattr(mod, "Dog")
attr = getattr(attr, "sleep")
# call
attr()
```
因此，在设计 LLGo Bindings 时，将类方法和静态方法转为函数：
> 为了接近 Python 的语法以及防止命名冲突，添加 `[ClassName]` 作为前缀。
```go
//go:linkname DogBark py.Dog.bark
func DogBark(msg *py.Object) *py.Object

//go:linkname DogSleep py.Dog.sleep
func DogSleep() *py.Object
```
用户使用方式：
```go
DogBark("Hello")
DogSleep()
```

### 多重继承
Python 示例：
```Python
class Parent1:
    def Method1(self):
        print(f"Parent1 Method1")

class Parent2:
    def Method2(self):
        print(f"Parent2 Method2")

class Child(Parent1, Parent2):
    def __init__(self, name):
        self._name = name
```
由于内存布局的问题，多重继承的效果并不能通过嵌入多个结构体来实现：

```go
type Child struct {
	Parent1
	Parent2
}
```
但在 Python 中，仍可以通过子类的符号链接到继承自父类的方法：
```Python
attr = getattr(mod, "Child")
attr = getattr(attr, "Method1")
# call
attr(child)
```

因此对于对于多继承的支持，一种可行的 LLGo Bindings 的设计方案是将子类定义为嵌入了 `py.Object` 对象的结构体，将继承到的所有父类方法都转为结构体方法：

```go
type Child struct {
	py.Object
}

//go:linkname (*Child).Method1 py.Child.Method1
func (c *Child) Method1() *py.Object
//go:linkname (*Child).Method2 py.Child.Method2
func (c *Child) Method2() *py.Object
```
**缺点**：无法查看继承关系，方法声明存在大量的代码冗余。
