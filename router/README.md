自定义路由处理器 准备移除

    调用方式：
        1. router.AddCompile(`auth/(\w+)`, "auth/$1")
        2. router.AddCompile(`auth/(\w+)`, "auth/$1", []middleware.Middleware{ //绑定中间件
	    	    &middleware.CorsMiddle{},
             }...)
        3. router.AddCompileGroup("checkLogin", map[string]string{ //按路由分组 绑定中间件
                `sysApi/(\w+)`: "sysApi/$1",
                `admin/(\w+)`:  "admin/$1",
            }, []middleware.Middleware{
                &middleware.CorsMiddle{},
                &middleware.LMiddle{},
            })

    注意：
        如果使用了自定义路由，最好也同时接管掉其他路由，这样可以避免自定义路由和默认路由同时存在
            // 过滤其他路由
            router.AddCompile(`\w.*`, "welcome/index")
        
        
