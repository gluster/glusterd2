Vagrant.configure("2") do |config|
  config.ssh.username = 'user'
  config.ssh.password = 'password'
  (1..4).each do |i|
    config.vm.define "consul#{i}" do |arch|
      arch.vm.hostname = "consul#{i}"
      arch.vm.provider "docker" do |d|
        d.image = 'kshlm/glusterd2-dev:latest'
        d.name = "consul#{i}"
        d.has_ssh = true
        d.remains_running = true
        d.volumes = [Dir.pwd+':/home/user/go']
        d.create_args = ['--privileged']
      end
    end
  end
end
