Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$myNumbersCollection = 1..5

if($myNumbersCollection -contains 3)
{
    "collection contains 3"
}
else
{
    "collection doesn't contain 3"
}
