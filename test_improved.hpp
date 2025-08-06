#pragma once

/**
    * @brief Namespace containing the core logic and data structures of the Test application.
    */
namespace TestNamespace {
    class TestClass {
    public:
        void publicMethod();
        int publicField;
    private:
        void privateMethod();
        int privateField_;
    };
    
    void globalFunction(int param);
    extern int globalVariable;
}
