package main

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// LLVMCompiler holds the LLVM compilation state
type LLVMCompiler struct {
	args      *CompilerArgs
	module    *ir.Module
	builder   *ir.Block
	currentFn *ir.Func
	locals    map[string]value.Value // local variables (alloca)
	globals   map[string]value.Value // global variables
	functions map[string]*ir.Func    // functions
	strings   []*ir.Global           // string constants
	labelID   int                    // counter for labels
	labels    map[string]*ir.Block   // named labels for goto
}

// NewLLVMCompiler creates a new LLVM compiler
func NewLLVMCompiler(args *CompilerArgs) *LLVMCompiler {
	return &LLVMCompiler{
		args:      args,
		module:    ir.NewModule(),
		locals:    make(map[string]value.Value),
		globals:   make(map[string]value.Value),
		functions: make(map[string]*ir.Func),
		strings:   make([]*ir.Global, 0),
	}
}

// GetModule returns the LLVM module
func (c *LLVMCompiler) GetModule() *ir.Module {
	return c.module
}

// WordType returns the B word type (i64)
func (c *LLVMCompiler) WordType() *types.IntType {
	return types.I64
}

// WordPtrType returns pointer to B word type
func (c *LLVMCompiler) WordPtrType() *types.PointerType {
	return types.NewPointer(c.WordType())
}

// DeclareGlobal declares a global variable
func (c *LLVMCompiler) DeclareGlobal(name string, init constant.Constant) *ir.Global {
	if init == nil {
		init = constant.NewInt(c.WordType(), 0)
	}
	global := c.module.NewGlobalDef(name, init)
	c.globals[name] = global
	return global
}

// DeclareGlobalArray declares a global array (vector)
// In B, arrays work as follows:
//   - name[N] creates N+1 words
//   - First word contains address of second word (pointer to data)
//   - Accessing name gives you the address of the first word
//   - Accessing name[i] loads the pointer and indexes into it
func (c *LLVMCompiler) DeclareGlobalArray(name string, size int64, init []constant.Constant) *ir.Global {
	arraySize := size + 1 // +1 for the pointer storage
	elemType := c.WordType()
	arrayType := types.NewArray(uint64(arraySize), elemType)

	// Initialize data elements
	var initVals []constant.Constant

	// First element: will be initialized with pointer to second element
	// We use a special constant expression: getelementptr to get address of element [1]
	// For now, use 0 and fix it with a global constructor or leave for linker
	// Actually, LLVM allows constant expressions for this
	initVals = append(initVals, constant.NewInt(elemType, 0)) // Placeholder, will be fixed below

	for i := int64(0); i < size; i++ {
		if init != nil && i < int64(len(init)) {
			initVals = append(initVals, init[i])
		} else {
			initVals = append(initVals, constant.NewInt(elemType, 0))
		}
	}

	global := c.module.NewGlobalDef(name, constant.NewArray(arrayType, initVals...))

	// Now fix the first element to point to the second element
	// Use constant GEP expression to get address of element [1]
	dataPtr := constant.NewGetElementPtr(arrayType, global,
		constant.NewInt(types.I64, 0),
		constant.NewInt(types.I64, 1))
	ptrAsInt := constant.NewPtrToInt(dataPtr, elemType)

	// Update the global initialization
	initVals[0] = ptrAsInt
	global.Init = constant.NewArray(arrayType, initVals...)

	c.globals[name] = global
	return global
}

// DeclareFunction declares a function
func (c *LLVMCompiler) DeclareFunction(name string, paramNames []string) *ir.Func {
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
func (c *LLVMCompiler) GetOrDeclareFunction(name string) *ir.Func {
	if fn, ok := c.functions[name]; ok {
		return fn
	}

	// Check if it was declared as extrn (exists in globals)
	// If so, DO NOT remove it from globals - it's a function pointer variable
	// Return nil to signal caller should handle it as indirect call
	if _, ok := c.globals[name]; ok {
		// It's an extrn variable (function pointer) - keep it in globals
		return nil
	}

	// Not in globals, not in functions → auto-declare as external variadic function
	// This handles undefined names used as functions (like write(), printf())
	fn := c.module.NewFunc(name, c.WordType())
	fn.Sig.Variadic = true
	c.functions[name] = fn
	return fn
}

// StartFunction starts building a function body
func (c *LLVMCompiler) StartFunction(fn *ir.Func) {
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
func (c *LLVMCompiler) EndFunction() {
	// If the current block doesn't have a terminator, add a default return
	if c.builder != nil && c.builder.Term == nil {
		c.builder.NewRet(constant.NewInt(c.WordType(), 0))
	}
	c.currentFn = nil
	c.builder = nil
	c.locals = make(map[string]value.Value)
	c.labels = make(map[string]*ir.Block)
}

// DeclareLocal allocates a local variable
func (c *LLVMCompiler) DeclareLocal(name string) value.Value {
	alloca := c.builder.NewAlloca(c.WordType())
	c.locals[name] = alloca
	return alloca
}

// DeclareLocalArray allocates a local array
// In B, arrays work as follows:
//   - array[N] allocates N+1 words
//   - First word contains pointer to second word (where data starts)
//   - This allows array[-1] to get the original pointer
func (c *LLVMCompiler) DeclareLocalArray(name string, size int64) value.Value {
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

// LoadValue loads a value (handles both locals and globals)
func (c *LLVMCompiler) LoadValue(name string) (value.Value, error) {
	// Check locals first
	if val, ok := c.locals[name]; ok {
		return c.builder.NewLoad(c.WordType(), val), nil
	}

	// Check globals
	if val, ok := c.globals[name]; ok {
		return c.builder.NewLoad(c.WordType(), val), nil
	}

	// Check if it's a function (return as pointer)
	if fn, ok := c.functions[name]; ok {
		return fn, nil
	}

	// Not found - will be declared later if it's a function call
	return nil, fmt.Errorf("undefined identifier '%s'", name)
}

// GetAddress gets the address of a variable (for lvalue operations)
// Returns nil if not found (will be handled as function call)
func (c *LLVMCompiler) GetAddress(name string) (value.Value, bool) {
	// Check locals first
	if val, ok := c.locals[name]; ok {
		return val, true
	}

	// Check functions before globals
	// This allows extrn-declared names to become functions if called
	if fn, ok := c.functions[name]; ok {
		return fn, true
	}

	// Check globals
	if val, ok := c.globals[name]; ok {
		return val, true
	}

	return nil, false
}

// StoreValue stores a value to a variable
func (c *LLVMCompiler) StoreValue(name string, val value.Value) error {
	addr, found := c.GetAddress(name)
	if !found {
		return fmt.Errorf("undefined identifier '%s'", name)
	}
	c.builder.NewStore(val, addr)
	return nil
}

// CreateStringConstant creates a global string constant
func (c *LLVMCompiler) CreateStringConstant(str string) *ir.Global {
	// Create a byte array for the string
	strBytes := []byte(str)
	strBytes = append(strBytes, 0) // null terminator

	// Create constant array
	charType := types.I8
	arrayType := types.NewArray(uint64(len(strBytes)), charType)

	var bytes []constant.Constant
	for _, b := range strBytes {
		bytes = append(bytes, constant.NewInt(charType, int64(b)))
	}

	strConst := constant.NewArray(arrayType, bytes...)
	global := c.module.NewGlobalDef(fmt.Sprintf(".str.%d", len(c.strings)), strConst)
	global.Immutable = true
	c.strings = append(c.strings, global)

	return global
}

// NewBlock creates a new basic block
func (c *LLVMCompiler) NewBlock(name string) *ir.Block {
	if name == "" {
		name = fmt.Sprintf("bb%d", c.labelID)
		c.labelID++
	}
	block := c.currentFn.NewBlock(name)
	return block
}

// SetInsertPoint sets the current insertion point
func (c *LLVMCompiler) SetInsertPoint(block *ir.Block) {
	c.builder = block
}

// GetInsertBlock returns the current insertion block
func (c *LLVMCompiler) GetInsertBlock() *ir.Block {
	return c.builder
}

// GetOrCreateLabel gets an existing label block or creates a new one
func (c *LLVMCompiler) GetOrCreateLabel(name string) *ir.Block {
	if block, ok := c.labels[name]; ok {
		return block
	}
	block := c.currentFn.NewBlock(name)
	c.labels[name] = block
	return block
}
