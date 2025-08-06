class TestSpan {
public:
    constexpr size_type size() const noexcept { return storage_.size; }
    constexpr size_type size_bytes() const noexcept { return size() * sizeof(element_type); }
    TCB_SPAN_NODISCARD constexpr bool empty() const noexcept { return size() == 0; }
    constexpr pointer data() const noexcept { return storage_.ptr; }
    constexpr iterator begin() const noexcept { return data(); }
    constexpr iterator end() const noexcept { return data() + size(); }
};
