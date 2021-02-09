# Unit Test
For Satellite, the specific plugin may have some common dependencies. So we provide a global test initializer to init the dependencies.


```
import (
   	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
   )
```

