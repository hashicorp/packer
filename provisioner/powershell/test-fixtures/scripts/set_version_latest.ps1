# Test fixture is a modified version of the example found at
# https://www.powershellmagazine.com/2012/10/23/pstip-set-strictmode-why-should-you-care/

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$myNumbersCollection = 1..5
if($myNumbersCollection -contains 3) {
    "collection contains 3"
}
else {
    "collection doesn't contain 3"
}