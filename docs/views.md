

# Views

## *a priori* 

At definition time, it is apparent:

- The possible permutations (note plural) of required parameters to support execution.
- Optional parameters.
- View schema:
    - `openapi` schema.
    - Relational schema.

## Runtime

The runtime representation of views must support:

- Views can be aliased as per tables.
- View columns can be aliased in the same way as table columns (even and **especially** those that are aliased inside the view itself).

## Ideation

- StackQL views DDL stored in some special stackql table designated for this purpose.
    - Physical table name such as `__iql__.views`.
    - Views need not exist until the `SELECT ... FROM <view>` portion of the query is executed.
      This is advantagesous on RDBMS systems where view creation will fail if physical tables do not exist.
    - We may need a layer of indirection for views to execute, wrt table names containing generation ID.
      Simplest option is input table name.
- SQL view definitions (translated to physical tables) are stored in the RDBMS.
    - This implies that even quite early in analysis, it must be known that a view is being referenced.
    - Some part of the namespace must be reserved for these views; configurable using existing regex / template namespacing?
    - Quite possibly some specialised object(s) or extension of the `table` interface stages are used for view analysis and parameter routing.
- Once analysis is complete:
    - Acquistion occurs as normal through primitive DAG.
    - Selection phase uses physical views.


## Subqueries

Some aspects of subquery analysis and execution will be similar to views, but not all.  What are the considerations for view implementation in the short term such that subsequent subquery implmentation is expedited and natural.

To be continued...


## Some example views

```sql
select 'ap-southeast-2' AS region, VolumeId, Encrypted, Size from aws.ec2.volumes where region = 'ap-southeast-2' UNION ALL SELECT 'ap-southeast-1' AS region, VolumeId, Encrypted, Size from aws.ec2.volumes where region = 'ap-southeast-1' UNION ALL SELECT 'ap-northeast-1' AS region, VolumeId, Encrypted, Size from aws.ec2.volumes where region = 'ap-northeast-1'  order by Size asc ;
```


```
|----------------|-----------------------|-----------|------|
|   aws_region   |       VolumeId        | Encrypted | Size |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-010ebc90228bcf0ce | false     |   12 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-032ad1e92c3d0a859 | false     |   12 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0cc0000bdb9e4ff58 | false     |   12 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0b0f03f1645bc77c9 | false     |   11 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0d156811b942d2df6 | false     |   11 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-018698380dda474b8 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-019d09408445b909e | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-044b158931786d740 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-05571d676274bb988 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0aaae07f47db9b77a | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0b1a6d50495b9d082 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0b946880bb06c7e70 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-northeast-1 | vol-0cbaa88dd7a0bfac9 | false     |   10 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-035b92417e36e74a1 | true      |   10 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-00a887ec28053d034 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-00b703f7df3a02a1a | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-010bde051e6ecd7f4 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-01366caacc6ec2d00 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-044caf3e728a30f87 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-049ee07b31aff451a | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0566b487a9485ac1f | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0616400eff26f3cc4 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0736baafbf8fa806a | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-079c6d13f241017c5 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-09e9b9a8ae37ab091 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0e424d83220df5b3b | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0f662d3776cc4b769 | false     |    8 |
|----------------|-----------------------|-----------|------|
| us-east-2      | vol-077905c95be09ddb0 | false     |    8 |
|----------------|-----------------------|-----------|------|
| us-east-2      | vol-08cbf143e71098fdd | false     |    8 |
|----------------|-----------------------|-----------|------|
| us-east-2      | vol-0afed29aa8f3f6966 | false     |    8 |
|----------------|-----------------------|-----------|------|
| ap-southeast-1 | vol-0a83b24ec495866e9 | false     |    7 |
|----------------|-----------------------|-----------|------|
```