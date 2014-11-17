#!/usr/bin/env ruby

# Usage
# cat log | time.rb

regex = /\[time\] ((?<total>\d+|\d+\.\d+)(?<total_units>us|ms|s)?) \(processing ((?<proc>\d+|\d+\.\d+)(?<proc_units>us|ms|s)?), requests ((?<req>\d+|\d+\.\d+)(?<req_units>us|ms|s)?)\)/
times = {}

begin
  ARGF.each_line do |line|
    m = regex.match(line)
    next if m.nil?
  
    [:total, :proc, :req].each do |category|
      time = m[category].to_f
      units = m[:"#{category}_units"]
      if units == "us"
        time /= 1000
      elsif units == "s"
        time *= 1000
      elsif units == "ms"
      elsif units.nil? && time == 0
      else
        puts "What are these units? #{units}"
        puts units.inspect
        exit(1)
      end
    
      times[category] ||= []
      times[category] << time
    end
  end
rescue Interrupt
end

[:total, :proc, :req].each do |category|
  puts category
  arr = times[category] || []
  puts "n: #{arr.size}"
  puts "avg: #{arr.inject(:+).to_f / arr.size}"
  puts "min: #{arr.min}"
  puts "max: #{arr.max}"
  puts ""
end
