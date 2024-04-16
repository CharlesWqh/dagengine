# DAG DSL
这里主要描述如何配置DAG的配置；

## 图集合 
每个文件中可以配置多个图，换言之，一个文件就是一组图的集合， 集合名称就是文件名；   
图集和有自己的配置，一般作用与整个图集和， 这里分别描述：

### **config_setting**
`config_setting`代表运行期的变量设置，可设置多个； 其格式如下：
```toml
[[config_setting]]
name = "with_exp_1000"
cond = "Recall1==1000"
```
`name`为变量名， 变量值为`cond`为表达式执行结果， 这里只支持`bool`结果，即表达式只能是返回bool值的表达式；  
变量在每次执行一个图前判断设置

## 图
图为一组顶点的集合， 除了`name`外无其它子属性；
```toml
[[graph]]
name = "auto_graph"
[[graph.vertex]]
# ...
```

## 顶点
顶点目前有两种类型，`算子`和`子图`， `算子`类型顶点又有两种子类型`常规算子` 和 `条件算子`：
- 子图顶点 （必须有`cluster`和`graph`两个属性）
- 算子顶点
    - 常规算子顶点 （必须有`processor`两个属性）
    - 条件算子顶点（必须有`cond`以及`if`和`else`两者至少一个的属性）

### 常规算子顶点
例如：
```toml
[[graph.vertex]]
processor = "phase0"
args = { abc = "default", xyz = 1010, hello=3.14, is_valid=true }
```

### 条件算子顶点
例如：
```toml
[[graph.vertex]]
id = "test_34old" # 算子id，大多数情况无需设置，存在歧义时需要设置 
cond = 'user_type=="34old"' # 条件算子表达式
if = ["subgraph_invoke"] # 条件算子true后继顶点
else = ["phase2"] # 条件算子false后继顶点
```
### 子图调用顶点
例如：
```toml
[[graph.vertex]]
id = "subgraph_invoke" # 子图调用算子id
cluster = "." # 子图集合名
graph = "sub_graph3" # 子图名
```

所有顶点有几个基本属性，以下分别说明：
### **id**
顶点的唯一ID（在单一的图中）； 在满足以下的情况下不用配置：
- 算子类型顶点， 且该算子在图中只出现一次；
- 或者该顶点未直接以及间接的被其它顶点在配置上直接依赖

未配置的情况下，DAG执行引擎会自动分配一个ID， 分配ID的规则如下：
- 若算子类型顶点， 则生成的ID = 算子processor name
- 其它情况下，按计数器int自增

### **args**
`args`代表传入给算子的运行参数，格式是toml的table格式，例如下面例子代表四种参数值的设置：
```toml
[[graph.vertex]]
processor = "phase0"
args = { abc = "default", xyz = 1010, hello=3.14, is_valid=true }
```

### **expect**
`expect`代表该顶点运行前的判断，其值为一个表达式， 若表达式值为true，则该顶点会运行，否则不会， 例如：
 ```toml
[[graph.vertex]]
# 若外部参数的user_type=="34old"， 该顶点才会运行
expect = '$user_type=="34old"'
processor = "phase3"
id = "phase3_2"
```

### **expect_config**
`expect_config`也代表该顶点运行前的判断，不同的是这里的配置是前文的`config_setting`的一个变量， 例如：
```toml
[[graph.vertex]]
# 若with_exp_1000为true， 该顶点才会运行
expect_config = "with_exp_1000"
processor = "phase3"
id = "phase3_2"
```

### **start**
`start`强制标识该顶点为起始顶点，绝大多数情况下不用配置，只有以下情况需要配置：   
**全图只有一个顶点**  
例如：
```toml
[[graph]]
name = "graph0"
[[graph.vertex]]
processor = "phase3"
start = true
```

### **select_args**
`select_args`代表条件选择参数，运行时会根据当时的变量取值选择一个合适的参数，一般还需要和`args`作为默认参数协作， 其中`match`应该是前文`config_setting`的一个变量名
```toml
[[graph.vertex]]
processor = "phase0"
select_args = [
    { match = "with_exp_1000", args = { abc = "hello1", xyz = "aaa" } },
    { match = "with_exp_1001", args = { abc = "hello2", xyz = "bbb" } },
    { match = "with_exp_1002", args = { abc = "hello3", xyz = "ccc" } },
]
# select未命中时的默认参数
args = { abc = "default", xyz = "zzz" }
```

### **successor/successor_on_ok/successor_on_err**
配置流程驱动时需要配置，`successor`含义为当前顶点访问完毕（无论是否执行，成功/失败），后继的顶点ID列表， 例如：
```toml
[[graph]]
name = "sub_graph2" # DAG图名  
[[graph.vertex]] # 顶点  
processor = "phase0" # 顶点算子，与子图定义/条件算子三选一
#id = "phase0"       # 算子id，大多数情况无需设置，存在歧义时需要设置; 这里默认id等于processor名
successor = ["test_34old", "compute_sth"] # 顶点后继顶点
```
三种`successor`的解释：
- successor, 当前顶点访问完毕（无论是否执行，成功/失败），后继的节点ID列表
- successor_on_ok, 当前顶点访问成功，后继的节点ID列表
- successor_on_err, 当前顶点访问失败，后继的节点ID列表

### **deps/deps_on_ok/deps_on_err**
配置流程驱动时需要配置，`deps`含义为当前顶点的前驱顶点ID列表，无论前驱节点的访问结果， 例如：
```toml
[[graph]]
name = "sub_graph2" # DAG图名  
[[graph.vertex]] # 顶点  
processor = "phase0" # 顶点算子，与子图定义/条件算子三选一
#id = "phase0"       # 算子id，大多数情况无需设置，存在歧义时需要设置; 这里默认id等于processor名
deps = ["pre0", "pre01"] # 顶点前驱顶点
```
三种`deps`的解释：
- deps, 无论前驱顶点是否执行，成功/失败，前驱顶点ID列表
- deps_on_ok, 前驱顶点访问成功情况下的顶点ID列表
- deps_on_err, 前驱顶点访问失败情况下的顶点ID列表

### **if/else**
`if/else`仅仅在条件顶点下配置，用于代表不同条件值下的后继执行顶点列表，例如：
```toml
[[graph.vertex]]
id = "test_34old" # 算子id，大多数情况无需设置，存在歧义时需要设置 
cond = 'user_type=="34old"' # 条件算子表达式
if = ["subgraph_invoke"] # 条件算子true后继顶点
else = ["phase2"] # 条件算子false后继顶点
```
### **input/output**
`input`代表顶点的输入，`output`代表顶点的输入; 仅在顶点类型为`常规算子`的情况下起作用： 由于DI的存在，大多数情况下并不需要配置，但也存在以下情况需要设置：
- 算子定义的input/output的名称+类型， 在整个图中不唯一；这里需要手动配置唯一的id
- 需要定义input的部分属性
    - 当input的数据为其它图算子的输出，需要定义属性`extern=true`
    - 当input为map类型，期望聚合指定的输出数据， 需要定义属性`aggregate=["id1", "id2"]`

一些示例：
```toml
[[graph.vertex]]
processor = "phase3"
id = "phase3_0"     # 显式设置ID
output = [{ field = "v100", id = "m0" }]
[[graph.vertex]]
expect = 'user_type=="12new"'   # 运行依赖条件
processor = "phase3"
id = "phase3_1"     # 显式设置ID
output = [{ field = "v100", id = "m1" }]   # 显式设置输出数据ID
[[graph.vertex]] 
expect = 'user_type=="34old"'   # 运行依赖条件
processor = "phase3"   # 显式设置ID
id = "phase3_2"
output = [{ field = "v100", id = "m2" }]   # 显式设置输出数据ID
[[graph.vertex]]
processor = "phase4"
input = [{ field = "v100", aggregate = ["m0", "m1", "m2"] }]   # 汇总依赖数据ID

```

`input/output` id可以配置为`$`变量形式，如：
```toml
[[graph.vertex]]
start = true
processor = "recall_merge"
input=[{field="r1", id="$input_name", extern = true}, {field="r2", extern = true},{field="r3", extern = true}]
```
变量值从执行的上下文参数中获取， 例如以下子图调用中的`$input_name`会被解释为`xyz`
```toml
[[graph.vertex]]
id = "recall_1"
cluster = "recall_test.toml"
graph = "recall_1"
args = {input_name="xyz"}
```


## 常见场景配置

### 多层实验
目前TAB上支持多层流量实验， 通常在一个较大的服务（召回、混排，排序等）上，存在一个服务对应多层流量，我们这里的使用经验是：
- 每层流量用一个固定图集和配置， 例如： `recall1.dag`对应召回的一层流量上所有实验
- 一般每个图代表一个实验流量， 例如： `recall1.dag`中的`exp_1001` 代表实验id为1001的实验
- 每层流量一般有个类似main函数的入口图用于dispatch到不同情况的图实现上；

例如RPC入口的图实现（一般固定）,代表固定的不同层流量协作关系:
```toml
[[graph]]
name = "main_entry" # DAG图名  
[[graph.vertex]] # 顶点  
id = "recall_0" # 子图调用算子id
cluster = "recall0.dag" # 子图集合名
graph = "main_entry"    # 子图名
[[graph.vertex]] # 顶点  
id = "recall_1" # 子图调用算子id
cluster = "recall1.dag" # 子图集合名
graph = "main_entry"    # 子图名
[[graph.vertex]] # 顶点  
id = "recall_2" # 子图调用算子id
cluster = "recall2.dag" # 子图集合名
graph = "main_entry"    # 子图名
[[graph.vertex]] # 顶点  
id = "recall_merge" # 子图调用算子id
cluster = "recall_merge.dag" # 子图集合名
graph = "main_entry"    # 子图名
deps = ["recall_0", "recall_1", "recall_2", "recall_3"]  
```
对应每层流量，如果大致流程固定，可以用一个固定的graph实现表示，新增流量实验就是新增顶点或者修改`select_args`参数, 这里不展开；  
这种实现对于较大规模频繁实验不太友好，所有开发都集中在一个graph上修改上线，冲突/维护的成本较高；

这里也设计有一种引入单独一个dispatch graph的实现：每层流量一般有个类似main函数的入口, 用于实现按实验/降级等标识的dispatch逻辑：
```toml
[[graph]]
name = "main_entry" # DAG图名  
[[graph.vertex]] # 顶点  
expect_config = "with_exp_10001"
id = "exp_10001" # 子图调用算子id
cluster = "recall0.dag" # 子图集合名
graph = "exp_10001"    # 子图名
[[graph.vertex]] # 顶点  
expect_config = "with_exp_10002"
id = "exp_10002" # 子图调用算子id
cluster = "recall0.dag" # 子图集合名
graph = "exp_10002"    # 子图名
[[graph.vertex]] 
id = "default" 
cluster = "recall0.dag" # 子图集合名
graph = "default"    # 子图名
des_on_err = ["exp_10001", "exp_10002"]  # 没有执行相关顶点情况下执行default顶点
```

### 多路召回汇总
假设一个召回算子设计被实现为通用实现，输出召回结果，行为则基于参数有不同的行为； 那么我们可以配置一个多路召回+merge如下：
```toml
[[graph]]
name = "main_entry" # DAG图名  
[[graph.vertex]]
processor = "common_recall" # 通用召回算子
id = "common_recall_0"     # 显式设置ID
args = {count = 100, model_id = 100}
output = [{ field = "recall_result", id = "r0" }]
[[graph.vertex]]
processor = "common_recall"
id = "common_recall_1"     # 显式设置ID
args = {count = 110, model_id = 101}
output = [{ field = "recall_result", id = "r1" }]
[[graph.vertex]] 
expect = 'user_type=="34old"'   # 运行依赖条件
processor = "common_recall"
id = "common_recall_1"     # 显式设置ID
args = {count = 110, model_id = 102}
output = [{ field = "recall_result", id = "r2" }]
[[graph.vertex]]
processor = "common_merge" # merge算子
input = [{ field = "recall_map", aggregate = ["r0", "r1", "r2"] }]   # 汇总依赖数据ID
```








