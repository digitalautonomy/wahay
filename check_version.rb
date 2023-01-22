#!/usr/bin/env ruby

left = ARGV[0]
right = ARGV[1]

leftC = left.split(/\./).map(&:to_i)
rightC = right.split(/\./).map(&:to_i)

fill = [leftC.length, rightC.length].max

([nil] * fill).zip(leftC, rightC) do |_, l, r|
  l = l || 0
  r = r || 0
  if l < r
    puts "lt"
    exit
  end
  if l > r
    puts "gt"
    exit
  end
end

puts "eq"
