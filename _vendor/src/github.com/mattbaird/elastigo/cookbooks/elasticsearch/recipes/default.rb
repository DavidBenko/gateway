#
# Cookbook Name:: elasticsearch
# Recipe:: default
# Author:: Matthew Baird <mattbaird@gmail.com>
#
# Copyright 2008-2012, Matthew Baird
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

user "#{node[:elasticsearch][:es_user]}" do
  system true
  shell "/bin/sh"
end


#make directories
directory node[:elasticsearch][:es_home] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:es_home]}') }
end

directory node[:elasticsearch][:config_dir] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:config_dir]}') }
end

directory node[:elasticsearch][:work_dir] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:work_dir]}') }
end

directory node[:elasticsearch][:log_dir] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:log_dir]}') }
end

directory node[:elasticsearch][:data_dir] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:data_dir]}') }
end

directory node[:elasticsearch][:mongo_plugin_dir] do
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0755"
  recursive true
  not_if { ::File.directory?('#{node[:elasticsearch][:mongo_plugin_dir]}') }
end

#move jars for mongo river
cookbook_file "#{node[:elasticsearch][:mongo_plugin_dir]}/mongo-2.9.1.jar" do
  source "mongo-2.9.1.jar"
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0555"
  notifies :reload, 'service[elasticsearch]'
end

cookbook_file "#{node[:elasticsearch][:mongo_plugin_dir]}/elasticsearch-river-mongodb-1.4.0-SNAPSHOT.jar" do
  source "elasticsearch-river-mongodb-1.4.0-SNAPSHOT.jar"
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0555"
  notifies :reload, 'service[elasticsearch]'
end

cookbook_file "#{node[:elasticsearch][:mongo_plugin_dir]}/elasticsearch-mapper-attachments-1.7.0-SNAPSHOT.jar" do
  source "elasticsearch-mapper-attachments-1.7.0-SNAPSHOT.jar"
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0555"
  notifies :reload, 'service[elasticsearch]'
end


include_recipe "elasticsearch::init"

service "elasticsearch" do
  supports :start => true, :stop => true, :restart => true
  provider Chef::Provider::Service::Upstart
end

# download a binary release, if new
remote_file "/tmp/elasticsearch-#{node[:elasticsearch][:version]}.tar.gz" do
  source "http://download.elasticsearch.org/elasticsearch/elasticsearch/elasticsearch-#{node[:elasticsearch][:version]}.tar.gz"
  not_if "test -f /tmp/elasticsearch-#{node[:elasticsearch][:version]}.tar.gz"
end

#Extract it, if it doesn't exist
execute "extract" do
  command "tar zxf /tmp/elasticsearch-#{node[:elasticsearch][:version]}.tar.gz"
  cwd "#{node[:elasticsearch][:extract]}"
  not_if do ::File.directory?("#{node[:elasticsearch][:es_home]}/bin") end
end

#register the service
template node[:elasticsearch][:config] do
  source "elasticsearch.conf.erb"
  owner "#{node[:elasticsearch][:es_user]}"
  group "#{node[:elasticsearch][:es_user]}"
  mode "0644"
  backup false
  notifies :restart, resources(:service => "elasticsearch"), :delayed
end

#start the service
service "elasticsearch" do
  action [:enable, :start]
end