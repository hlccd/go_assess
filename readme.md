# readme
###因为上传图片过于麻烦,所以在发往邮箱的邮件里附带了pdf格式的readme
## 任务要求

```
- 登录、注册、保持登录状态          (已完成)
- 开始一局游戏          (已完成)
  - 随机匹配模式，此模式下只会匹配到同样选择了随机匹配模式的人          (已完成)
  - 自己创建房间，等待他人加入。双方都进入房间后，需要同时选择准备才会开始游戏。          (已完成)
  - 查看所有以创建的房间，选择一个房间进入          (已完成)
- 游戏对局
  - 一局定胜负（多局游戏比较麻烦但不是很难，不是考察的重点）          (已完成)
  - 开始后，双方需要在二十秒钟之内决定自己出剪刀、石头或布。          (已完成)
    - 结果有三种，胜利、失败或平局          (已完成)
    - 到二十秒钟时，如果双方都不出，视为平局；如果一方出了一方没出，出了的那一方获胜          (已完成)
  - 通知玩家游戏结果          (已完成)
  - 结束后，回到等待准备状态          (已完成)
- 游戏记录
  - 保存每一局游戏的记录          (已完成)
  - 自己可以查看自己的游戏记录          (已完成)
加分项
- 优秀的项目结构、详细的注释、优雅的代码或易懂的说明文档          (应该说明文档不会很难看)
- 高并发          (应该算是完成了)
- 匹配旗鼓相当的对手          (已完成)
- 断线重连          (已完成)
- 在同一个房间内可以发送文字消息          (已完成)
- 观战模式（可以看到两位玩家是否准备、游戏结果、发送消息等）          (已完成)
- 部署你的项目（使用nginx、docker等）          (已完成)
- 客户端（能用就行）          (已完成)
- 其他酷炫的功能
```

## 剪刀石头布

1.账号系统

2.游戏等待or匹配

3.对局

4.历史对局记录

### 数据表

##### 账号表

| 名称     | 类型        | 描述                                  |
| -------- | ----------- | ------------------------------------- |
| id       | bigint      | 游戏账号唯一ID,主键,自99999开始自增长 |
| password | varchar(30) | 游戏账号密码                          |
| rate     | float       | 胜率,用于拓展匹配等级                 |
| win      | bigint      | 总获胜次数                            |
| total    | bigint      | 总场数                                |

##### Session结构

| 名称      | 类型  | 描述                                     |
| --------- | ----- | ---------------------------------------- |
| sid       | int64 | session 唯一ID号                         |
| id        | int64 | 用户ID号                                 |
| rid       | int64 | 房间号,为0则不在房间内,匹配仍有房间号    |
| state     | int   | 状态,0为已登录,1为在房间内               |
| timestamp | int64 | 最近操作的时间戳,后续拓展自动清除session |

##### monitor结构

**当且仅当账号进入房间后才开始监听**

| 名称 | 类型     | 描述                     |
| ---- | -------- | ------------------------ |
| conn | net.Conn | 与客户端的链接           |
| id   | int64    | 该链接客户端的id         |
| sid  | int64    | 该链接客户端账号的sid    |
| rid  | int64    | 该链接客户端所处的房间号 |

##### message结构

| 名称 | 类型   | 描述                 |
| ---- | ------ | -------------------- |
| Mes  | string | 消息主体             |
| Id   | int64  | 消息发送者的id       |
| Sid  | int64  | 消息发送者的Sid      |
| Rid  | int64  | 可接收该消息的房间号 |



##### 历史对局记录表

**每个用户独立一张表,每张表记录与用户的对局对手和输赢情况**

| 名称        | 类型   | 描述                                                |
| ----------- | ------ | --------------------------------------------------- |
| opponent_id | bitint | 对手ID号                                            |
| win         | int    | 输赢情况,1为我方赢,0为我方输,求和得到我方总获胜次数 |
| my_gesture  | int    | 我方出的手势,0为没出,1为剪刀,2为石头,3为布          |
| op_gesture  | int    | 对方出的手势,0为没出,1为剪刀,2为石头,3为布          |
| timestamp   | bigint | 时间戳,用以记录时间                                 |

##### Record结构

历史对局记录结构,用于读取数据库后做暂存

| 名称      | 类型  | 描述                     |
| --------- | ----- | ------------------------ |
| Opponent  | int64 | 对手ID号                 |
| Win       | int   | 是否获胜,1为胜,0为负或平 |
| MyGesture | int   | 该局我的手势             |
| OpGesture | int   | 该局对方的手势           |
| Timestamp | int64 | 此局结束的时间戳         |

##### Room结构

房间分四类,正常创建房间,低胜率匹配房间,中胜率匹配房间,高胜率匹配房间

| 名称     | 类型  | 描述                                                         |
| -------- | ----- | ------------------------------------------------------------ |
| room_id  | int64 | 房间id号                                                     |
| player1  | int64 | 一号玩家ID号                                                 |
| player2  | int64 | 二号玩家ID号                                                 |
| prepare1 | int   | 一号玩家准备情况,1为已准备,0为未准备,2为游戏开始但没出手势,3为剪刀,4为石头,5为布 |
| prepare2 | int   | 二号玩家准备情况,1为已准备,0为未准备,2为游戏开始但没出手势,3为剪刀,4为石头,5为布 |

### 账号系统

#### 账号注册

**请求方式:POST**

**URL:http://47.108.217.244:2995/account/register**

##### 输入参数

| key      | 描述         |
| -------- | ------------ |
| password | 新账号的密码 |

##### 输出参数

| key      | 描述       |
| -------- | ---------- |
| code     | 执行码     |
| id       | 用户id号   |
| password | 用户的密码 |
| rate     | 用户的胜率 |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\1_1.png)

#### 账号登陆

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/account/login

##### 输入参数

| key      | 描述       |
| -------- | ---------- |
| id       | 账号的id   |
| password | 账号的密码 |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\1_2.png)

#### 账号登出

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/account/logout

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\1_3.png)

#### 历史对局

##### 请求方式:GET

##### URL:http://47.108.217.244:2995/account/record

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key    | 描述                                              |
| ------ | ------------------------------------------------- |
| code   | 执行码                                            |
| id     | 用户id号                                          |
| sid    | 用户绑定的session的id                             |
| record | 历史对局记录数组,具体结构见数据表中Record结构描述 |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\1_4.png)

![](C:\Users\94976\Desktop\go_assess\readme\1_5.png)

### 房间系统

#### 创建房间

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/room/insert

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |
| rid  | 新建的房间编号        |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_1.png)

#### 可加入房间列表

##### 请求方式:GET

##### URL:http://47.108.217.244:2995/room/list

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key   | 描述                                    |
| ----- | --------------------------------------- |
| code  | 执行码                                  |
| id    | 用户id号                                |
| sid   | 用户绑定的session的id                   |
| rooms | 房间结构体数组,结构描述见数据表Room结构 |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_2.png)

#### 所有存在的房间列表

##### 请求方式:GET

##### URL:http://47.108.217.244:2995/room/all

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key   | 描述                                    |
| ----- | --------------------------------------- |
| code  | 执行码                                  |
| id    | 用户id号                                |
| sid   | 用户绑定的session的id                   |
| rooms | 房间结构体数组,结构描述见数据表Room结构 |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_3.png)

#### 进入房间

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/room/enter

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |
| rid  | 房间id号              |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |
| rid  | 房间编号              |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_4.png)

#### 观战

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/room/view

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |
| rid  | 房间id号              |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |
| rid  | 房间编号              |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_5.png)

#### 离开房间

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/room/leave

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |
| rid  | 房间id号              |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |
| rid  | 房间编号              |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_6.png)

#### 匹配

##### 请求方式:POST

##### URL:http://47.108.217.244:2995/random/match

##### 输入参数

| key  | 描述                  |
| ---- | --------------------- |
| id   | 账号的id              |
| sid  | 用户绑定的session的id |

##### 输出参数

| key  | 描述                  |
| ---- | --------------------- |
| code | 执行码                |
| id   | 用户id号              |
| sid  | 用户绑定的session的id |
| rid  | 新建的房间编号        |

##### 示例

![](C:\Users\94976\Desktop\go_assess\readme\2_7.png)

### 以上是以http协议实现的内容,当进入房间后,需要使用socket开启监听

**因随机匹配和新建房间的实际形式都是处于房间中,故以下仅演示双方竞赛时有一人观战的过程,随机匹配同理**

**使用socket监听的文件名为"customer.go",可自行下载使用**

**以下以99999为一号玩家,100000为2号玩家,100001为旁观者为例**

![](C:\Users\94976\Desktop\go_assess\readme\3_1.png)

![3_2](C:\Users\94976\Desktop\go_assess\readme\3_2.png)

![3_3](C:\Users\94976\Desktop\go_assess\readme\3_3.png)

**其中,99999和100000输出1表示以准备,当两方都准备后游戏开始,处于房间的任何人都可以自由发言,除了"1","3","4","5"作为保留指令外,其他可以让全房间看到**
