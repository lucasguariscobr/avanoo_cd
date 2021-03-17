Vagrant.configure('2') do |config|
  config.vm.box = "ubuntu/bionic64"
  config.ssh.forward_agent = true
  config.vm.network 'forwarded_port', guest: 5500, host: 5500
  config.vm.synced_folder '.', '/home/vagrant/go/src/cd'

  config.vm.provider 'virtualbox' do |virtualbox|
    virtualbox.memory = 4096
    virtualbox.cpus = 2
    virtualbox.customize ['modifyvm', :id, '--uartmode1', 'disconnected']
  end
end