osm
===

osm(Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。

以前是使用MyBatis开发java服务端，它的sql mapping很灵活，把sql独立出来，程序通过输入与输出来完成所有的数据库操作。

osm就是对MyBatis的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。sql xml 格式如下：
  <?xml version="1.0" encoding="utf-8"?>
  <osm>
   <select id="selectUsers" result="structs">
     SELECT id,email
     FROM user
     {{if ne .Email ""}} where email=#{Email} {{end}}
     order by id
   </select>
  </osm>
