package main

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// Compiler holds the LLVM compilation state
type Compiler struct {
	args      *CompileOptions
	module    *ir.Module
	builder   *ir.Block
	currentFn *ir.Func
	locals    map[string]value.Value // local variables (alloca)
	globals   map[string]value.Value // global variables
	functions map[string]*ir.Func    // functions
	strings   []*ir.Global           // string constants
	stringID  int                    // unique id for string constants
	labelID   int                    // counter for labels
	labels    map[string]*ir.Block   // named labels for goto
}

// NewCompiler creates a new compiler structure
func NewCompiler(args *CompileOptions) *Compiler {
	return &Compiler{
		args:      args,
		module:    ir.NewModule(),
		locals:    make(map[string]value.Value),
		globals:   make(map[string]value.Value),
		functions: make(map[string]*ir.Func),
		strings:   make([]*ir.Global, 0),
		stringID:  0,
	}
}

// findFuncByName returns the function with the given name from the module, or nil
func (c *Compiler) findFuncByName(name string) *ir.Func {
	for _, f := range c.module.Funcs {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// removeFuncByName removes the function with the given name from the module if present
func (c *Compiler) removeFuncByName(name string) bool {
	for i, f := range c.module.Funcs {
		if f.Name() == name {
			c.module.Funcs = append(c.module.Funcs[:i], c.module.Funcs[i+1:]...)
			return true
		}
	}
	return false
}

// findGlobalByName returns the global with the given name from the module, or nil
func (c *Compiler) findGlobalByName(name string) *ir.Global {
	for _, g := range c.module.Globals {
		if g.Name() == name {
			return g
		}
	}
	return nil
}

// removeGlobalByName removes the global with the given name from the module if present
func (c *Compiler) removeGlobalByName(name string) bool {
	for i, g := range c.module.Globals {
		if g.Name() == name {
			c.module.Globals = append(c.module.Globals[:i], c.module.Globals[i+1:]...)
			return true
		}
	}
	return false
}

// GetModule returns the LLVM module
func (c *Compiler) GetModule() *ir.Module {
	return c.module
}

// WordType returns the B word type (i64)
func (c *Compiler) WordType() *types.IntType {
	return types.I64
}

// WordPtrType returns pointer to B word type
func (c *Compiler) WordPtrType() *types.PointerType {
	return types.NewPointer(c.WordType())
}

// DeclareGlobal declares a global variable
func (c *Compiler) DeclareGlobal(name string, init constant.Constant) *ir.Global {
	var global *ir.Global
	if init == nil {
		// If no initializer, create zero-initialized global
		global = c.module.NewGlobalDef(name, constant.NewInt(c.WordType(), 0))
	} else {
		global = c.module.NewGlobalDef(name, init)
	}
	c.globals[name] = global
	return global
}

// DeclareGlobalWithMultipleValues declares a global scalar with multiple initialization values
// This allocates N consecutive words but treats it as a scalar (not an array)
// Example: c -345, 'foo', "bar";  allocates 3 words starting at &c
func (c *Compiler) DeclareGlobalWithMultipleValues(name string, values []constant.Constant) *ir.Global {
	// Create an array type to hold all values
	arrayType := types.NewArray(uint64(len(values)), c.WordType())

	// Create constant array with all values
	arrayInit := constant.NewArray(arrayType, values...)

	// Create the global - it's a scalar name but backed by array storage
	global := c.module.NewGlobalDef(name, arrayInit)
	c.globals[name] = global
	return global
}

// DeclareGlobalArray declares a global array (vector)
// In B, arrays work as follows:
//   - name[N] creates N+1 words
//   - First word contains address of second word (pointer to data)
//   - Accessing name gives you the address of the first word
//   - Accessing name[i] loads the pointer and indexes into it
//
// For large arrays without initializers (size > 10), we use a two-global approach:
//   - name.data[N] = zeroinitializer (very compact in .ll files!)
//   - name[1] = [pointer to name.data]
//
// This dramatically reduces .ll file size for large zero-initialized arrays.
func (c *Compiler) DeclareGlobalArray(name string, size int64, init []constant.Constant) *ir.Global {
	// True B semantics: arrays are pointers to buffers.
	// Emit: name.data = [size x i64] <inits>
	//       name = i64 ptrtoint(&name.data[0])
	elemType := c.WordType()

	if size < 0 {
		size = 0
	}

	dataArrayType := types.NewArray(uint64(size), elemType)

	// Build data initializer values, padding with zeros
	dataVals := make([]constant.Constant, size)
	for i := int64(0); i < size; i++ {
		if init != nil && i < int64(len(init)) {
			dataVals[i] = init[i]
		} else {
			dataVals[i] = constant.NewInt(elemType, 0)
		}
	}

    dataInit := constant.NewArray(dataArrayType, dataVals...)
    dataGlobal := c.module.NewGlobalDef(name+".data", dataInit)

    // Compute pointer to first element
    zeroI64 := constant.NewInt(types.I64, 0)
    firstElemPtr := constant.NewGetElementPtr(dataArrayType, dataGlobal, zeroI64, zeroI64)

    // The array variable itself is a pointer to i64 holding the address
    // of the first element (avoid storing raw ptr-as-int for relocatability).
    global := c.module.NewGlobalDef(name, firstElemPtr)
	c.globals[name] = global
	return global
}

// DeclareFunction declares a function
func (c *Compiler) DeclareFunction(name string, paramNames []string) *ir.Func {
	// Remove any prior function with the same name from the module to enforce no-context
	c.removeFuncByName(name)

	// All B functions take i64 parameters and return i64
	params := make([]*ir.Param, len(paramNames))
	for i, pname := range paramNames {
		params[i] = ir.NewParam(pname, c.WordType())
	}

	fn := c.module.NewFunc(name, c.WordType(), params...)
	c.functions[name] = fn
	return fn
}

// GetOrDeclareFunction gets an existing function or declares it as external
//
// B language semantics:
//   - Undefined identifier used as function: auto-declare as external function
//   - 'extrn name' then 'name(...)': name is a function pointer variable (indirect call)
//
// Examples:
//
//	printf("hello");         → auto-declares printf as external function (direct call)
//	extrn printf; printf(); → printf is a variable holding function pointer (indirect call)
func (c *Compiler) GetOrDeclareFunction(name string) *ir.Func {
	if fn, ok := c.functions[name]; ok {
		return fn
	}

	// If declared as extrn variable in current context, treat as function pointer variable
	if c.findGlobalByName(name) != nil {
		return nil
	}

	// Enforce no-context: remove any prior function of the same name from the module
	c.removeFuncByName(name)

	// Auto-declare as external variadic function in current context
	fn := c.module.NewFunc(name, c.WordType())
	fn.Sig.Variadic = true
	c.functions[name] = fn
	return fn
}

// ClearTopLevelContext clears symbol tables that should not persist across
// top-level declarations while keeping the LLVM module intact.
func (c *Compiler) ClearTopLevelContext() {
	c.globals = make(map[string]value.Value)
	c.functions = make(map[string]*ir.Func)
	c.strings = make([]*ir.Global, 0)
}

// StartFunction starts building a function body
func (c *Compiler) StartFunction(fn *ir.Func) {
	c.currentFn = fn
	c.locals = make(map[string]value.Value)
	c.labels = make(map[string]*ir.Block)
	c.builder = fn.NewBlock("entry")

	// Allocate space for parameters
	for _, param := range fn.Params {
		alloca := c.builder.NewAlloca(c.WordType())
		c.builder.NewStore(param, alloca)
		c.locals[param.Name()] = alloca
	}
}

// EndFunction finalizes a function
func (c *Compiler) EndFunction() {
	// If the current block doesn't have a terminator, add a default return
	if c.builder != nil && c.builder.Term == nil {
		c.builder.NewRet(constant.NewInt(c.WordType(), 0))
	}
	c.currentFn = nil
	c.builder = nil
	c.locals = make(map[string]value.Value)
	c.labels = make(map[string]*ir.Block)
}

// DeclareLocal allocates a local variable and initializes it to 0
func (c *Compiler) DeclareLocal(name string) value.Value {
	alloca := c.builder.NewAlloca(c.WordType())
	// Initialize local variables to 0 (B language semantics)
	c.builder.NewStore(constant.NewInt(c.WordType(), 0), alloca)
	c.locals[name] = alloca
	return alloca
}

// DeclareLocalArray allocates a local array
// In B, arrays work as follows:
//   - array[N] allocates N+1 words
//   - First word contains pointer to second word (where data starts)
//   - This allows array[-1] to get the original pointer
//   - Empty arrays (size 0) default to size 1
func (c *Compiler) DeclareLocalArray(name string, size int64) value.Value {
	// Empty arrays default to size 1
	if size == 0 {
		size = 1
	}
	arraySize := size + 1 // +1 for pointer storage in first element
	arrayType := types.NewArray(uint64(arraySize), c.WordType())
	alloca := c.builder.NewAlloca(arrayType)

	// Get pointer to first data element (skip the pointer storage slot)
	// This is element [0][1] in the array
	firstElemPtr := c.builder.NewGetElementPtr(arrayType, alloca,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 1))

	// Convert to i64 and store in the first slot [0][0]
	ptrAsInt := c.builder.NewPtrToInt(firstElemPtr, c.WordType())
	firstSlotPtr := c.builder.NewGetElementPtr(arrayType, alloca,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 0))
	c.builder.NewStore(ptrAsInt, firstSlotPtr)

	// Store the array base address (to first slot which now contains the data pointer)
	c.locals[name] = firstSlotPtr
	return firstSlotPtr
}

// GetAddress gets the address of a variable (for lvalue operations)
// Returns nil if not found (will be handled as function call)
func (c *Compiler) GetAddress(name string) (value.Value, bool) {
	// Check locals first
	if val, ok := c.locals[name]; ok {
		return val, true
	}

	// Check module functions
	if fn := c.findFuncByName(name); fn != nil {
		return fn, true
	}

	// Check module globals
	if irGlobal := c.findGlobalByName(name); irGlobal != nil {
		// If the global is an array type (scalar with multiple values)
		if arrayType, ok := irGlobal.ContentType.(*types.ArrayType); ok {
			firstElem := c.builder.NewGetElementPtr(arrayType, irGlobal,
				constant.NewInt(types.I32, 0),
				constant.NewInt(types.I32, 0))
			return firstElem, true
		}
		return irGlobal, true
	}

	return nil, false
}

// CreateStringConstant creates a global string constant
func (c *Compiler) CreateStringConstant(str string) *ir.Global {
	// Create null-terminated string constant
	// CharArrayFromString doesn't include null terminator, so we need to add it manually
	strBytes := []byte(str)
	strBytes = append(strBytes, 0) // null terminator

	charType := types.I8
	arrayType := types.NewArray(uint64(len(strBytes)), charType)

	var bytes []constant.Constant
	for _, b := range strBytes {
		bytes = append(bytes, constant.NewInt(charType, int64(b)))
	}

	strConst := constant.NewArray(arrayType, bytes...)

	global := c.module.NewGlobalDef(fmt.Sprintf(".str.%d", c.stringID), strConst)
	global.Linkage = enum.LinkagePrivate
	global.UnnamedAddr = enum.UnnamedAddrUnnamedAddr
	global.Immutable = true
	c.strings = append(c.strings, global)
	c.stringID++

	return global
}

// NewBlock creates a new basic block
func (c *Compiler) NewBlock(name string) *ir.Block {
	if name == "" {
		name = fmt.Sprintf("bb%d", c.labelID)
		c.labelID++
	}
	block := c.currentFn.NewBlock(name)
	return block
}

// SetInsertPoint sets the current insertion point
func (c *Compiler) SetInsertPoint(block *ir.Block) {
	c.builder = block
}

// GetInsertBlock returns the current insertion block
func (c *Compiler) GetInsertBlock() *ir.Block {
	return c.builder
}

// GetOrCreateLabel gets an existing label block or creates a new one
func (c *Compiler) GetOrCreateLabel(name string) *ir.Block {
	if block, ok := c.labels[name]; ok {
		return block
	}
	block := c.currentFn.NewBlock(name)
	c.labels[name] = block
	return block
}
