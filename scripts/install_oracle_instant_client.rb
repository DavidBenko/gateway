#!/usr/bin/env ruby

require 'rbconfig'

def download_url
  "http://downloads.justapis.com/v5.1.0/oracle/#{os}_instant_client_12_1.tar.gz"
end

def os
  @os ||= (
    host_os = RbConfig::CONFIG['host_os']
    case host_os
    when /mswin|msys|mingw|cygwin|bccwin|wince|emc/
      :windows
    when /darwin|mac os/
      :osx
    when /linux/
      :linux
    when /solaris|bsd/
      :unix
    else
      raise Error, "unknown os: #{host_os.inspect}"
    end
  )
end

def append(file, line_to_append)
  `echo '#{line_to_append}' >> #{file}`
  raise "Failed to append #{line_to_append} to #{file}" unless $?.success?
end

if os == :osx
  INSTANT_CLIENT_DIR_LOCATION = ARGV[0]
  INSTANT_CLIENT_DIR_NAME="instantclient_12_1"
  INSTANT_CLIENT_DIR = "#{INSTANT_CLIENT_DIR_LOCATION}/#{INSTANT_CLIENT_DIR_NAME}"
  OCI_INC_DIR="#{INSTANT_CLIENT_DIR}/sdk/include"
  `mkdir -p #{INSTANT_CLIENT_DIR_LOCATION} && curl --silent #{download_url} | tar -C #{INSTANT_CLIENT_DIR_LOCATION} -zxv `
  raise "Failed to download client from #{download_url} into #{INSTANT_CLIENT_DIR}" unless $?.success?
  append(ARGV[1],"\nexport DYLD_LIBRARY_PATH=#{INSTANT_CLIENT_DIR}") unless ENV['DYLD_LIBRARY_PATH'] == INSTANT_CLIENT_DIR
  append(ARGV[1],"\nexport OCI_LIB_DIR=#{INSTANT_CLIENT_DIR}") unless ENV['OCI_LIB_DIR'] == INSTANT_CLIENT_DIR
  append(ARGV[1],"\nexport OCI_INC_DIR=#{OCI_INC_DIR}") unless ['OCI_INC_DIR'] == OCI_INC_DIR
end
