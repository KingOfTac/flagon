command {
  name = "lua-foo",
  description = "A Lua command that prints a message",

  args = {
    { name = "message", description = "The message to print" }
  },

  middleware = {
    function(ctx)
      ctx.log("info", "before build")
      ctx.next()
      ctx.log("info", "after build")
    end
  },

  handler = function(ctx)
    print("Hello from Lua!", ctx.args[1])
  end
}