绑定到http的处理器

解析uri

http返回信息剥离-默认处理格式
支持自定义http返回结构处理

    handle.SetCustomHttpReturn(&customReturn.CustomReturn{})

匹配路由

    [静态文件解析]

    [控制器方法匹配]

配合router生产路由规则

AddHandleStruct() 新增对外的绑定关系，自动获取struct对外的方法生成映射关系

默认使用structName[首字母小写] 方法第一个参数必须为github.com/solaa51/swagger/context *context.Context类型

> handle 绑定方法调用 绑定的就是函数的调用地址
>
> 这样的话，所有数据只能依靠参数传递ctx需负责所有参数处理
>
> 可导出的http对外方法，第一个参数必须为ctx
>
> Controller的成员变量，变成了类似静态的概念
>
> 变相去掉了new(struct)带来的内存分配，把所有的请求都映射到了具体的方法上
