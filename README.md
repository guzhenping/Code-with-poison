# Code-with-poison TiDB hackathon 2019
本项目旨在解决用户的SQL 质量参差不齐，迁移和使用过程中，引发集群资源紧张的问题。

# 背景
在手动迁移800+ 存量MySQL SQL -> TiDB的过程中，观察集群CPU和Mem消耗，找到了不少问题。包括
- 本身有索引，未用上
  - SQL中在where条件里使用了内置函数(cast/date/json)，导致优化器不能用索引
  - SQL中where使用与column类型不同的value，比如：varchar = int
  - SQL中使用IN语法，导致优化器错判使用Tablescan，忽略索引

- 细节问题
  - SQL中对bit类型的字段查询，使用值等于 '1' / '0' (char类型)
  - SQL中where条件用到varchar类型字段的值等于，MySQL可以大小写不敏感，TiDB大小写敏感
  - 无法兼容使用MySQL table: mysql.help_topic


以上问题，需要过滤和优化：
- 过滤：SQL中仅用到TableScan
- 过滤：SQL中仅用到TableScan，且Selection/TableReader的Count的行数大于1000w
- 优化：指定where条件中的字段，使用IndexScan
- 优化：提示添加索引，add index idx_xx(where条件中的字段)
- 优化：提示强制指定索引，use index( column_name)

# 优化案例：
表结构：
```
+----------------------+--------------+------+------+---------+----------------+
| Field                | Type         | Null | Key  | Default | Extra          |
+----------------------+--------------+------+------+---------+----------------+
| trip_id              | bigint(20)   | NO   | PRI  | NULL    | auto_increment |
| duration             | int(11)      | NO   |      | NULL    |                |
| start_date           | datetime     | YES  |      | NULL    |                |
| end_date             | datetime     | YES  |      | NULL    |                |
| start_station_number | int(11)      | YES  |      | NULL    |                |
| start_station        | varchar(255) | YES  |      | NULL    |                |
| end_station_number   | int(11)      | YES  | MUL  | NULL    |                |
| end_station          | varchar(255) | YES  |      | NULL    |                |
| bike_number          | varchar(255) | YES  |      | NULL    |                |
| member_type          | varchar(255) | YES  | MUL  | NULL    |                |
+----------------------+--------------+------+------+---------+----------------+
```

8个优化案例：
```
case1 : 本身存在索引 但是未使用

目标SQL:
SELECT * FROM bikeshare.trips WHERE member_type = 123;

explain 信息:
Selection_5 1137929 root eq(cast(bikeshare.trips.member_type), 123)
└─TableReader_7 1422412 root data:TableScan_6
  └─TableScan_6 1422412 cop table:trips, range:[-inf,+inf], keep order:false

过滤与优化结果:
[INFO] please use index1
===========================================



case2 : 该列不存在索引 可以加索引

目标SQL:
SELECT * FROM bikeshare.trips WHERE duration > 123;

explain 信息:
TableReader_7 1404493 root data:Selection_6
└─Selection_6 1404493 cop gt(bikeshare.trips.duration, 123)
  └─TableScan_5 1422412 cop table:trips, range:[-inf,+inf], keep order:false

过滤与优化结果:
[INFO] need add index
===========================================



case3: 正常 正在使用索引

目标SQL:
SELECT * FROM bikeshare.trips WHERE member_type = 'test';

explain 信息:
IndexLookUp_10 0 root 
├─IndexScan_8 0 cop table:trips, index:member_type, range:["test","test"], keep order:false
└─TableScan_9 0 cop table:trips, keep order:false

过滤与优化结果:
[INFO] Good, using IndexScan
===========================================



case4: 该列不存在索引 可以加索引

目标SQL:
SELECT * FROM bikeshare.trips WHERE start_station_number > 123;

explain 信息:
TableReader_7 1422412 root data:Selection_6
└─Selection_6 1422412 cop gt(bikeshare.trips.start_station_number, 123)
  └─TableScan_5 1422412 cop table:trips, range:[-inf,+inf], keep order:false

过滤与优化结果:
[INFO] need add index
===========================================



case5:tableScan 数据量正常

目标SQL:
SELECT * FROM bikeshare.trips WHERE member_type = 'test' union all SELECT * FROM trips WHERE start_station_number = 1

explain 信息:
Union_10 0 root 
├─IndexLookUp_17 0 root 
│ ├─IndexScan_15 0 cop table:trips, index:member_type, range:["test","test"], keep order:false
│ └─TableScan_16 0 cop table:trips, keep order:false
└─TableReader_21 0 root data:Selection_20
  └─Selection_20 0 cop eq(bikeshare.trips.start_station_number, 1)
    └─TableScan_19 1422412 cop table:trips, range:[-inf,+inf], keep order:false

模拟阈值 300000
过滤与优化结果:
[INFO] TableScan_16 scan data correct
===========================================



case6:局部 tableScan 数据量不正常

目标SQL:
SELECT * FROM bikeshare.trips WHERE member_type = 'test' union all SELECT * FROM trips WHERE start_station_number > 1

explain 信息:
Union_10 1422412 root 
├─IndexLookUp_17 0 root 
│ ├─IndexScan_15 0 cop table:trips, index:member_type, range:["test","test"], keep order:false
│ └─TableScan_16 0 cop table:trips, keep order:false
└─TableReader_21 1422412 root data:Selection_20
  └─Selection_20 1422412 cop gt(bikeshare.trips.start_station_number, 1)
    └─TableScan_19 1422412 cop table:trips, range:[-inf,+inf], keep order:false

模拟阈值 300000
过滤与优化结果:
[INFO] TableScan_16 scan data correct
[INFO] TableScan_19 scan data too much
===========================================



case7:未使用索引 使用优化策略 输出优化后SQL

目标SQL:
SELECT count(*) FROM bikeshare.trips WHERE cast(end_station_number as char) > '123';

explain 信息:
StreamAgg_10 1 root funcs:count(1)
└─Selection_14 1137929 root gt(cast(bikeshare.trips.end_station_number), "123")
  └─TableReader_16 1422412 root data:TableScan_15
    └─TableScan_15 1422412 cop table:trips, range:[-inf,+inf], keep order:false

过滤与优化结果:
[INFO] please use index: SELECT count(*) FROM bikeshare.trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';
===========================================



case8:索引使用正常

目标SQL:
SELECT count(*) FROM bikeshare.trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';

explain 信息:
StreamAgg_10 1 root funcs:count(1)
└─Selection_14 1137929 root gt(cast(bikeshare.trips.end_station_number), "123")
  └─IndexReader_16 1422412 root index:IndexScan_15
    └─IndexScan_15 1422412 cop table:trips, index:end_station_number, range:[NULL,+inf], keep order:false

过滤与优化结果:
[INFO] Good, using IndexScan
===========================================
···
