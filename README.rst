Structer
========

**Note**: This badly needed a lick of paint before Go Modules came onto the scene,
and should now be considered obsolete/unmaintained. Having said that, if you depend
on this and you run into trouble with it, please open an issue and I'll see if I
can help.

--

Structer is a tool for dismantling struct definitions to try to ease the agony
of code generation.

Structer requires Go 1.9.

It ties together `go/types <https://godoc.org/go/types>`_ and `go/ast
<https://godoc.org/go/ast>`_ to try to simplify recursively walking through a
type declaration, importing relevant pacakges and extracting required code and
documentation.

The API is very much a moving target at this point - the main use case is
supporting my alternative `msgp code generator
<https://github.com/shabbyrobe/msgpgen>`_ (which supplements `msgp
<https://github.com/tinylib/msgp>`_ with some sorely needed conveniences), but
any reports of other use cases that could be supported would be great!

Efficiency is not a major goal of this project - sewing together the disparate
and disjointed APIs of `go/types`, `go/ast` and `go/doc` is the overriding
priority.

The core of structer is the ``TypePackageSet``, which allows you to import
complete packages and their ASTs::

    tpset := structer.NewTypePackageSet()
    pkg, err := tpset.Import("path/to/pkg")

You can then recursively walk type definitions (importing external defs as you
like by calling back to ``TypePackageSet``) in order to generate code. Either
fully implement ``structer.TypeVisitor`` yourself, or just part of it using
``structer.PartialTypeVisitor``::

    ppv := &structer.PartialTypeVisitor{
        EnterStructFunc: func(s s structer.StructInfo) error {
            return nil
        },
        LeaveStructFunc: func(s structer.StructInfo) error {
            return nil
        },
        EnterFieldFunc: func(s s structer.StructInfo, field *types.Var, tag string) error {
            return nil
        },
        LeaveFieldFunc: func(s s structer.StructInfo, field *types.Var, tag string) error {
            return nil
        },
        VisitBasic(t *types.Basic) error {
            fmt.Printf("Found basic leaf node %s", t.String())
            return nil
        },
        VisitNamed(t *types.Named) error {
            fmt.Printf("Found named leaf node %s, importing", t.String())
            _, err := tpset.ImportNamed(t)
            return err
        },
    }

``Walk`` will visit every part of a compound type declaration and stop only at
``types.Basic`` or ``types.Named`` declarations.

``Walk`` shouldn't even have trouble with this crazy thing::

    type Pants struct {
        Foo struct {
            Bar map[struct{ X, Y int}]map[struct{ Y, Z int}]struct {
                Baz []*Pants
            }
        }
    }
    

Structer also allows you to extract all constant values across all imported
packages that have a certain type::

    tpset := structer.NewTypePackageSet()
    pkg, err := tpset.Import("path/to/pkg")
    consts, err := tpset.ExtractConstants(structer.NewTypeName("path/to/pkg", "MyEnum"), false)


Known limitations
-----------------

- API is very unstable
- Not enough tests yet
- Poor documentation
- ``TypeVisitor`` should possibly have  ``EnterMap`` and ``LeaveMap``.
  ``EnterMapKey`` and ``LeaveMapValue`` may be sufficient, but they're less clear.

