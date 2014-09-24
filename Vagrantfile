# - Drop this Vagrantfile into your GOPATH
# - Change into GOPATH and run vagrant up
#   - You'll need to have docker installed
#   - You may also need to provide the provider
#     option to the `vagrant up` command
#     `vagrant --provider=docker up`
#   - The first time you bring up the environment
#     use the --no-parallel argument. This will
#     allow docker to safely download the image
#     for the first time.
#     `vagrant up --no-parallel`
# - You can now ssh into the containers using
#   `vagrant ssh <name>` and work inside the
#   containers.

Vagrant.configure("2") do |config|
  config.ssh.username = 'user'
  config.ssh.password = 'password'
  (1..4).each do |i|
    config.vm.define "glusterd2-dev-#{i}" do |arch|
      arch.vm.hostname = "glusterd2-dev-#{i}"
      arch.vm.provider "docker" do |d|
        d.image = 'kshlm/glusterd2-dev:latest'
        d.name = "glusterd2-dev-#{i}"
        d.has_ssh = true
        d.remains_running = true
        d.volumes = [Dir.pwd+':/home/user/go']
        d.create_args = ['--privileged']
      end
    end
  end
end
