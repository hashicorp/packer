# copyright: 2018, The Authors

title "skip control"

control "skip-1.0" do
  impact 1.0
  title "skip control"
  desc "skip control to generate a 101 return code"
  only_if { 1 == 2 }
  describe file("/tmp") do
    it { should be_directory }
  end
end
