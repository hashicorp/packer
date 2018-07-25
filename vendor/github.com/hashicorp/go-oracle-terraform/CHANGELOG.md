## 0.6.6 (January 11, 2018)

* compute: Create and delete machine images [GH-101]

## 0.6.5 (January 8, 2018)

* compute: Orchestration failures should explicitly tell the user why it failed [GH-100]

## 0.6.4 (Decemeber 20, 2017) 

* compute: Added suspend functionality to orchestrated instances [GH-99]

## 0.6.3 (December 13, 2017)

* storage: Added remove header option to storage objects and containers [GH-96]

## 0.6.2 (November 28, 2017)

* client: Added a UserAgent to the Client [GH-98]

## 0.6.1 (Novemeber 26, 2017)

* compute: Added is_default_gateway to network attributes for instances [GH-97]


## 0.6.0 (November 10, 2017)

* compute: Added is_default_gateway to network attributes for instances [GH-90]

* compute: Added the orchestration resource, specifically for instance creation [GH-91]

## 0.5.1 (October 5, 2017)

* java: Fixed subscription_type field

## 0.5.0 (October 5, 2017)

* java: Added more fields to java service instance [GH-89]

## 0.4.0 (September 14, 2017)

* database: Add utility resources [GH-87]

* compute: Increase storage volume snapshot create timeout [GH-88]

## 0.3.4 (August 16, 2017)

* storage_volumes: Actually capture errors during a storage volume create ([#86](https://github.com/hashicorp/go-oracle-terraform/issues/86))

## 0.3.3 (August 10, 2017)

* Add `ExposedHeaders` to storage containers ([#85](https://github.com/hashicorp/go-oracle-terraform/issues/85))

* Fixed `AllowedOrigins` in storage containers ([#85](https://github.com/hashicorp/go-oracle-terraform/issues/85))

## 0.3.2 (August 7, 2017)

* Add `id` for storage objects ([#84](https://github.com/hashicorp/go-oracle-terraform/issues/84))

## 0.3.1 (August 7, 2017)

* Update tests for Database parameter changes ([#83](https://github.com/hashicorp/go-oracle-terraform/issues/83))

## 0.3.0 (August 7, 2017)
 
 * Add JaaS Service Instances ([#82](https://github.com/hashicorp/go-oracle-terraform/issues/82))
 
 * Add storage objects ([#81](https://github.com/hashicorp/go-oracle-terraform/issues/81))
 
## 0.2.0 (July 27, 2017)

 * service_instance: Switches yes/no strings to bool in input struct and then converts back to strings for ease of use on user end ([#80](https://github.com/hashicorp/go-oracle-terraform/issues/80))

## 0.1.9 (July 20, 2017)

 * service_instance: Update delete retry count ([#79](https://github.com/hashicorp/go-oracle-terraform/issues/79))
 
 * service_instance: Add additional fields ([#79](https://github.com/hashicorp/go-oracle-terraform/issues/79))

## 0.1.8 (July 19, 2017)

 * storage_volumes: Add SSD support ([#78](https://github.com/hashicorp/go-oracle-terraform/issues/78))

## 0.1.7 (July 19, 2017)

  * database: Adds the Oracle Database Cloud to the available sdks. ([#77](https://github.com/hashicorp/go-oracle-terraform/issues/77))
  
  * database: Adds Service Instances to the database sdk ([#77](https://github.com/hashicorp/go-oracle-terraform/issues/77))

## 0.1.6 (July 18, 2017)

 * opc: Add timeouts to instance and storage inputs ([#75](https://github.com/hashicorp/go-oracle-terraform/issues/75))

## 0.1.5 (July 5, 2017)

 * storage: User must pass in Storage URL to CRUD resources ([#74](https://github.com/hashicorp/go-oracle-terraform/issues/74))

## 0.1.4 (June 30, 2017)

 * opc: Fix infinite loop around auth token exceeding it's 25 minute duration. ([#73](https://github.com/hashicorp/go-oracle-terraform/issues/73))

## 0.1.3 (June 30, 2017)

  * opc: Add additional logs instance logs ([#72](https://github.com/hashicorp/go-oracle-terraform/issues/72))
  
  * opc: Increase instance creation and deletion timeout ([#72](https://github.com/hashicorp/go-oracle-terraform/issues/72))

## 0.1.2 (June 30, 2017)


FEATURES:

  * opc: Add image snapshots ([#67](https://github.com/hashicorp/go-oracle-terraform/issues/67))
  
  * storage: Storage containers have been added ([#70](https://github.com/hashicorp/go-oracle-terraform/issues/70))


IMPROVEMENTS: 
  
  * opc: Refactored client to be generic for multiple Oracle api endpoints ([#68](https://github.com/hashicorp/go-oracle-terraform/issues/68))
  
  * opc: Instance creation retries when an instance enters a deleted state ([#71](https://github.com/hashicorp/go-oracle-terraform/issues/71))
  
## 0.1.1 (May 31, 2017)

IMPROVEMENTS:

 * opc: Add max_retries capabilities ([#66](https://github.com/hashicorp/go-oracle-terraform/issues/66))
 
## 0.1.0 (May 25, 2017)

BACKWARDS INCOMPATIBILITIES / NOTES:

 * Initial Release of OPC SDK
