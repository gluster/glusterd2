# This only works if you have a single GOPATH
#
Vagrant.configure("2") do |config|
  (1..4).each do |i|
    config.vm.define "gd2-#{i}" do |arch|
      arch.vm.hostname = "gd2-#{i}"
      arch.vm.provider "docker" do |d|
        d.image = 'kshlm/glusterd2-dev:centos-latest'
        d.name = "gd2-#{i}"
        d.has_ssh = true
        d.remains_running = true
        d.volumes = [ENV["GOPATH"]+':/go']
        d.create_args = ['--privileged']
        # Use the below only if you have dnsdock setup
        d.create_args += ['--dns', '172.17.0.1']
      end
    end
  end
end
