#!/usr/bin/env ruby

require 'rbconfig'
require 'fileutils'

INSTANT_CLIENT_DIR = ARGV[0]

def download_url
  "http://downloads.justapis.com/v5.1.0/oracle/#{os}_#{architecture}_instant_client_12_1.tar.gz"
end

def os
  # Allow override of detected OS.
  if ARGV[2] || ENV['HOST_OS']
    @os = (ARGV[2] || ENV['HOST_OS']).to_sym
    puts "Using #{ARGV[2] || ENV['HOST_OS']} as OS type passed from commandline."
    return @os
  end
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

def architecture
  # Allow override of detected architecture.
  if ARGV[3] || ENV['HOST_ARCH']
    puts "Using #{ARGV[3] || ENV['HOST_ARCH']} as architecture type passed from commandline."
    return (ARGV[3] || ENV['HOST_ARCH']).to_i
  end
  1.size * 8
end

def create_config(template, new_first_line, destination_file)
  original_config = File.read(template)
  new_config = new_first_line
  new_config << original_config
  File.open(destination_file, 'w') {|f| f.puts new_config }
end

def do_install(extension)
 if File.symlink?(File.join(INSTANT_CLIENT_DIR,"libclntsh.#{extension}"))
    puts "Instant client appears to be already installed."
  else
    puts "Setting up instant client in #{INSTANT_CLIENT_DIR}"
    FileUtils::mkdir_p(INSTANT_CLIENT_DIR)
    `curl --silent #{download_url} | tar -zxv -C #{INSTANT_CLIENT_DIR} --strip 1`
    raise "Failed to download client from #{download_url} into #{INSTANT_CLIENT_DIR}" unless $?.success?
    prefix = "prefix=#{INSTANT_CLIENT_DIR}"
    oci8_pc = File.join(INSTANT_CLIENT_DIR,'oci8.pc')
    if (ENV['HOST_OS'] && ENV['HOST_OS'].eql?("windows"))
      create_config("#{ARGV[1]}.windows", prefix, oci8_pc)
    else
      create_config(ARGV[1], prefix, oci8_pc)
    end
    `cd #{INSTANT_CLIENT_DIR} && ln -s libclntsh.#{extension}.12.1 libclntsh.#{extension}`
    raise "Failed to create a symbolic link!" unless $?.success?
  end
end

if os == :osx
  do_install('dylib')
elsif os == :linux
  do_install('so')
elsif os == :windows
  do_install('dll')
else
  raise "Please implement me for #{os}!"
end
