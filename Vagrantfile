# This only works if you have a single GOPATH
#
Vagrant.configure("2") do |config|
  config.ssh.username = 'user'
  config.ssh.password = 'password'
  (1..4).each do |i|
    config.vm.define "gd2-#{i}" do |arch|
      arch.vm.hostname = "gd2-#{i}"
      arch.vm.provider "docker" do |d|
        d.image = 'kshlm/glusterd2-dev:latest'
        d.name = "gd2-#{i}"
        d.has_ssh = true
        d.remains_running = true
        d.volumes = [ENV["GOPATH"]+':/home/user/go']
        d.create_args = ['--privileged']
        # Use the below only if you have skydock and skydns setup. Change the domain (dev.docker) to match your setup
        #d.create_args = ['--privileged', '--dns', '172.17.42.1', '--dns-search', 'glusterd2-dev.dev.docker']
      end
    end
  end
end
