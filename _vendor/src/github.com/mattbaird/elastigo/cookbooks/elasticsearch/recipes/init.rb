template "/etc/init/elasticsearch.conf" do
  cookbook "elasticsearch"
  source "elasticsearch.conf.erb"
  mode 0644
  backup false
end
