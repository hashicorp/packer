# Change Log

## [v0.1.4](https://github.com/vultr/govultr/compare/v0.1.3..v0.1.4) (2019-07-14)
### Bug Fixes
* Fix panic on request failure [#20](https://github.com/vultr/govultr/pull/20)

## [v0.1.3](https://github.com/vultr/govultr/compare/v0.1.2..v0.1.3) (2019-06-13)
### Features
* added `GetVc2zList` to Plans to retrieve `high-frequency compute` plans [#13](https://github.com/vultr/govultr/pull/13)

### Breaking Changes
* Renamed all variables named `vpsID` to `instanceID` [#14](https://github.com/vultr/govultr/pull/14)
* Server
    * Renamed Server struct field `VpsID` to `InstanceID` [#14](https://github.com/vultr/govultr/pull/14)
* Plans
    * Renamed Plan struct field `VpsID` to `PlanID` [#14](https://github.com/vultr/govultr/pull/14)
    * Renamed BareMetalPlan struct field `BareMetalID` to `PlanID` [#14](https://github.com/vultr/govultr/pull/14)
    * Renamed VCPlan struct field `VpsID` to `PlanID` [#14](https://github.com/vultr/govultr/pull/14)
    * Renamed Plan struct field `VCPUCount` to `vCPUs` [#13](https://github.com/vultr/govultr/pull/13)
    * Renamed BareMetalPlan struct field `CPUCount` to `CPUs` [#13](https://github.com/vultr/govultr/pull/13)
    * Renamed VCPlan struct field `VCPUCount` to `vCPUs` [#13](https://github.com/vultr/govultr/pull/13)
    * Renamed VCPlan struct field `Cost` to `Price` [#13](https://github.com/vultr/govultr/pull/13)

## [v0.1.2](https://github.com/vultr/govultr/compare/v0.1.1..v0.1.2) (2019-05-29)
### Fixes
* Fixed Server Option `NotifyActivate` bug that ignored a `false` value
* Fixed Bare Metal Server Option `UserData` to be based64encoded 
### Breaking Changes
* Renamed all methods named `GetList` to `List`
* Renamed all methods named `Destroy` to `Delete`
* Server Service
    * Renamed `GetListByLabel` to `ListByLabel`
    * Renamed `GetListByMainIP` to `ListByMainIP`
    * Renamed `GetListByTag` to `ListByTag`
* Bare Metal Server Service
    * Renamed `GetListByLabel` to `ListByLabel`
    * Renamed `GetListByMainIP` to `ListByMainIP`
    * Renamed `GetListByTag` to `ListByTag`

## [v0.1.1](https://github.com/vultr/govultr/compare/v0.1.0..v0.1.1) (2019-05-20)
### Features
* add `SnapshotID` to ServerOptions as an option during server creation
* bumped default RateLimit from `.2` to `.6` seconds
### Breaking Changes
* Iso
  * Renamed all instances of `Iso` to `ISO`.  
* BlockStorage
  * Renamed `Cost` to `CostPerMonth`
  * Renamed `Size` to `SizeGB` 
* BareMetal & Server 
  * Change `SSHKeyID` to `SSHKeyIDs` which are now `[]string` instead of `string`
  * Renamed `OS` to `Os`    

## v0.1.0 (2019-05-10)
### Features
* Initial release
* Supports all available API endpoints that Vultr has to offer